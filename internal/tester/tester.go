package tester

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

const tag = "tester"

type Config struct {
	Attempts        int
	Timeout         time.Duration
	ConcurrentTests int
	TestDelay       time.Duration
}

func DefaultConfig() Config {
	return Config{
		Attempts:        3,
		Timeout:         5 * time.Second,
		ConcurrentTests: 10,
		TestDelay:       100 * time.Millisecond,
	}
}

type Tester struct {
	config  Config
	limiter chan struct{}
}

func NewTester(config Config) *Tester {
	return &Tester{
		config:  config,
		limiter: make(chan struct{}, config.ConcurrentTests),
	}
}

func (t *Tester) Test(cfg *common.NormalizedConfig) *common.PingResult {
	addr := fmt.Sprintf("%s:%d", cfg.Server, cfg.Port)
	logger.Infof(tag, "testing %q -> %s (%d attempts, timeout=%v)",
		cfg.Name, addr, t.config.Attempts, t.config.Timeout)

	result := &common.PingResult{
		Config:   cfg,
		Attempts: t.config.Attempts,
	}

	var latencies []int64
	var successfulAttempts int

	for i := 0; i < t.config.Attempts; i++ {
		if i > 0 {
			time.Sleep(t.config.TestDelay)
		}

		latency, err := t.pingOnce(addr)
		if err != nil {
			logger.Debugf(tag, "  [%s] attempt %d/%d: FAILED (%v)",
				cfg.Name, i+1, t.config.Attempts, err)
			continue
		}

		logger.Debugf(tag, "  [%s] attempt %d/%d: OK (%d ms)",
			cfg.Name, i+1, t.config.Attempts, latency)
		latencies = append(latencies, latency)
		successfulAttempts++
	}

	result.Successes = successfulAttempts
	result.Reachable = successfulAttempts > 0

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

	logger.PingReport(tag, cfg.Name, result.Attempts, result.Successes,
		result.AvgMs, result.MinMs, result.MaxMs)

	return result
}

func (t *Tester) TestAll(configs []*common.NormalizedConfig) []*common.PingResult {
	logger.Infof(tag, "testing %d config(s) with concurrency=%d",
		len(configs), t.config.ConcurrentTests)

	results := make([]*common.PingResult, len(configs))
	var wg sync.WaitGroup

	for i, cfg := range configs {
		wg.Add(1)

		t.limiter <- struct{}{}

		go func(index int, config *common.NormalizedConfig) {
			defer wg.Done()
			defer func() { <-t.limiter }()

			results[index] = t.Test(config)
		}(i, cfg)
	}

	wg.Wait()

	reachable := 0
	totalLatency := int64(0)
	latencyCount := 0

	for _, r := range results {
		if r.Reachable {
			reachable++
			if r.AvgMs > 0 {
				totalLatency += r.AvgMs
				latencyCount++
			}
		}
	}

	avgLatency := int64(0)
	if latencyCount > 0 {
		avgLatency = totalLatency / int64(latencyCount)
	}

	logger.Infof(tag, "test complete: %d/%d reachable (%.1f%%), avg latency: %d ms",
		reachable, len(results),
		float64(reachable)*100.0/float64(len(results)),
		avgLatency)

	return results
}

func (t *Tester) pingOnce(addr string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.config.Timeout)
	defer cancel()

	start := time.Now()

	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return 0, err
	}
	conn.Close()

	latency := time.Since(start).Milliseconds()
	return latency, nil
}

func RankResults(results []*common.PingResult) []*common.PingResult {
	ranked := make([]*common.PingResult, len(results))
	copy(ranked, results)

	for i := 0; i < len(ranked); i++ {
		for j := i + 1; j < len(ranked); j++ {
			if shouldSwap(ranked[i], ranked[j]) {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}

	return ranked
}

func shouldSwap(a, b *common.PingResult) bool {
	if a.Reachable != b.Reachable {
		return !a.Reachable
	}
	if !a.Reachable {
		return false
	}

	if a.LossPercent != b.LossPercent {
		return a.LossPercent > b.LossPercent
	}

	return a.AvgMs > b.AvgMs
}

func Test(cfg *common.NormalizedConfig, testCfg Config) *common.PingResult {
	tester := NewTester(testCfg)
	return tester.Test(cfg)
}

func TestAll(configs []*common.NormalizedConfig, testCfg Config) []*common.PingResult {
	tester := NewTester(testCfg)
	return tester.TestAll(configs)
}
