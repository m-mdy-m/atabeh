package tester

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

const tag = "tester"

type Config struct {
	Attempts int
	Timeout  time.Duration
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		Attempts: 3,
		Timeout:  5 * time.Second,
	}
}

func Test(cfg *common.NormalizedConfig, testCfg Config) *common.PingResult {
	addr := fmt.Sprintf("%s:%d", cfg.Server, cfg.Port)
	logger.Infof(tag, "testing %q -> %s (%d attempts, timeout=%v)",
		cfg.Name, addr, testCfg.Attempts, testCfg.Timeout)

	result := &common.PingResult{
		Config:   cfg,
		Attempts: testCfg.Attempts,
	}

	var latencies []int64

	for i := 0; i < testCfg.Attempts; i++ {
		latency, err := pingOnce(addr, testCfg.Timeout)
		if err != nil {
			logger.Debugf(tag, "  attempt %d/%d: FAILED (%v)", i+1, testCfg.Attempts, err)
			continue
		}
		logger.Debugf(tag, "  attempt %d/%d: OK (%d ms)", i+1, testCfg.Attempts, latency)
		latencies = append(latencies, latency)
		result.Successes++
	}

	result.Reachable = result.Successes > 0

	if len(latencies) > 0 {
		result.MinMs = latencies[0]
		result.MaxMs = latencies[0]
		var sum int64
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

	// verbose report via logger
	logger.PingReport(tag, cfg.Name, result.Attempts, result.Successes,
		result.AvgMs, result.MinMs, result.MaxMs)

	return result
}

// TestAll runs Test on each config and returns all results
func TestAll(configs []*common.NormalizedConfig, testCfg Config) []*common.PingResult {
	logger.Infof(tag, "testing %d config(s)", len(configs))

	var results []*common.PingResult
	for _, cfg := range configs {
		r := Test(cfg, testCfg)
		results = append(results, r)
	}

	// summary
	reachable := 0
	for _, r := range results {
		if r.Reachable {
			reachable++
		}
	}
	logger.Infof(tag, "test complete: %d/%d reachable", reachable, len(results))

	return results
}

// pingOnce attempts a single TCP connection and returns latency in ms.
// Uses context with timeout so we don't hang forever.
func pingOnce(addr string, timeout time.Duration) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return 0, err
	}
	conn.Close()

	latency := time.Since(start).Milliseconds()
	return latency, nil
}
