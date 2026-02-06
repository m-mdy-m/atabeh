package tester

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

const tag = "tester"

func New(cfg Config) *Tester {
	return &Tester{
		config:  cfg,
		limiter: make(chan struct{}, cfg.ConcurrentTests),
	}
}

func Test(cfg *common.NormalizedConfig, testCfg Config) *Result {
	t := New(testCfg)
	return t.test(cfg)
}

func TestAll(configs []*common.NormalizedConfig, testCfg Config) []*Result {
	t := New(testCfg)
	return t.testAll(configs)
}

func (t *Tester) test(cfg *common.NormalizedConfig) *Result {
	addr := fmt.Sprintf("%s:%d", cfg.Server, cfg.Port)
	logger.Infof(tag, "testing %q -> %s (%d attempts)", cfg.Name, addr, t.config.Attempts)

	result := &Result{
		Config:   cfg,
		Attempts: t.config.Attempts,
	}

	var latencies []int64
	var failureReasons []string

	for i := 0; i < t.config.Attempts; i++ {
		if i > 0 && t.config.TestDelay > 0 {
			time.Sleep(t.config.TestDelay)
		}

		latency, failTag, err := t.pingOnce(addr, cfg)
		if err != nil {
			logger.Debugf(tag, "[%s] attempt %d/%d: FAIL (%v)",
				cfg.Name, i+1, t.config.Attempts, err)
			if failTag != "" {
				failureReasons = append(failureReasons, failTag)
			}
			continue
		}

		logger.Debugf(tag, "[%s] attempt %d/%d: OK (%d ms)",
			cfg.Name, i+1, t.config.Attempts, latency)
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

		if IsSuspiciouslyFast(result.AvgMs) {
			if !validateRealConnection(addr, cfg) {
				logger.Warnf(tag, "[%s] FAKE PING detected", cfg.Name)
				result.Reachable = false
				result.Successes = 0
				result.FailureTag = "fake-ping"
			}
		}

		if t.config.StabilityWindow > 0 && result.Reachable {
			result.StabilityScore = t.testStability(addr, cfg)
			logger.Debugf(tag, "[%s] stability score: %.2f", cfg.Name, result.StabilityScore)

			if result.StabilityScore < 0.5 {
				logger.Warnf(tag, "[%s] UNSTABLE connection", cfg.Name)
				result.Reachable = false
				result.FailureTag = "unstable"
			}
		}
	} else if t.config.TagFailures && len(failureReasons) > 0 {

		result.FailureTag = MostCommon(failureReasons)
	}

	return result
}

func (t *Tester) testAll(configs []*common.NormalizedConfig) []*Result {
	logger.Infof(tag, "testing %d configs (concurrent=%d)",
		len(configs), t.config.ConcurrentTests)

	results := make([]*Result, len(configs))
	var wg sync.WaitGroup

	for i, cfg := range configs {
		wg.Add(1)
		t.limiter <- struct{}{}

		go func(idx int, c *common.NormalizedConfig) {
			defer wg.Done()
			defer func() { <-t.limiter }()
			results[idx] = t.test(c)
		}(i, cfg)
	}

	wg.Wait()

	alive := 0
	for _, r := range results {
		if r.Reachable {
			alive++
		}
	}

	logger.Infof(tag, "complete: %d/%d reachable (%.1f%%)",
		alive, len(results), float64(alive)*100.0/float64(len(results)))

	return results
}

func (t *Tester) pingOnce(addr string, cfg *common.NormalizedConfig) (int64, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.config.Timeout)
	defer cancel()

	start := time.Now()
	conn, failTag, err := DialWithContext(ctx, "tcp", addr, cfg)
	if err != nil {
		return 0, failTag, err
	}
	defer conn.Close()

	latency := time.Since(start).Milliseconds()
	return latency, "", nil
}

func (t *Tester) testStability(addr string, cfg *common.NormalizedConfig) float64 {
	ctx, cancel := context.WithTimeout(context.Background(), t.config.StabilityWindow)
	defer cancel()

	start := time.Now()
	successCount := 0
	totalTests := 0

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if totalTests == 0 {
				return 0.0
			}
			return float64(successCount) / float64(totalTests)

		case <-ticker.C:
			totalTests++
			_, _, err := t.pingOnce(addr, cfg)
			if err == nil {
				successCount++
			}

			if time.Since(start) >= t.config.StabilityWindow {
				if totalTests == 0 {
					return 0.0
				}
				return float64(successCount) / float64(totalTests)
			}
		}
	}
}
