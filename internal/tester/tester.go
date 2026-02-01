package tester

import (
	"context"
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

const tag = "tester"

type Config struct {
	Attempts int
	Timeout  time.Duration
}

func DefaultConfig() Config {
	return Config{Attempts: 3, Timeout: 5 * time.Second}
}

func Test(cfg *common.NormalizedConfig, testCfg Config) *common.PingResult {
	addr := fmt.Sprintf("%s:%d", cfg.Server, cfg.Port)
	logger.Infof(tag, "testing %q → %s (%d attempts, timeout=%v)",
		cfg.Name, addr, testCfg.Attempts, testCfg.Timeout)

	result := &common.PingResult{
		Config:   cfg,
		Attempts: testCfg.Attempts,
	}

	var latencies []int64
	for i := 0; i < testCfg.Attempts; i++ {
		lat, err := pingOnce(addr, testCfg.Timeout)
		if err != nil {
			logger.Debugf(tag, "  attempt %d/%d: FAIL (%v)", i+1, testCfg.Attempts, err)
			continue
		}
		logger.Debugf(tag, "  attempt %d/%d: OK   (%d ms)", i+1, testCfg.Attempts, lat)
		latencies = append(latencies, lat)
		result.Successes++
	}

	result.Reachable = result.Successes > 0

	if len(latencies) > 0 {
		var sum int64
		result.MinMs = latencies[0]
		result.MaxMs = latencies[0]
		for _, l := range latencies {
			sum += l
			if l < result.MinMs {
				result.MinMs = l
			}
			if l > result.MaxMs {
				result.MaxMs = l
			}
		}
		result.AvgMs = sum / int64(len(latencies))
	}

	if result.Attempts > 0 {
		result.LossPercent = ((result.Attempts - result.Successes) * 100) / result.Attempts
	}

	logger.PingReport(tag, cfg.Name, result.Attempts, result.Successes,
		result.AvgMs, result.MinMs, result.MaxMs)

	return result
}

func TestAll(configs []*common.NormalizedConfig, testCfg Config) []*common.PingResult {
	logger.Infof(tag, "testing %d config(s)", len(configs))

	results := make([]*common.PingResult, 0, len(configs))
	for _, cfg := range configs {
		results = append(results, Test(cfg, testCfg))
	}

	// Print the final summary
	logger.SummaryReport(results)
	return results
}

func RankResults(results []*common.PingResult) []*common.PingResult {
	ranked := make([]*common.PingResult, len(results))
	copy(ranked, results)

	sort.SliceStable(ranked, func(i, j int) bool {
		ri, rj := ranked[i], ranked[j]
		if ri.Reachable != rj.Reachable {
			return ri.Reachable
		}
		return ri.AvgMs < rj.AvgMs
	})
	return ranked
}

func pingOnce(addr string, timeout time.Duration) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return 0, err
	}
	conn.Close()
	return time.Since(start).Milliseconds(), nil
}
