package tester

import (
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
)

type Config struct {
	Attempts        int
	Timeout         time.Duration
	ConcurrentTests int
	TestDelay       time.Duration
	StabilityWindow time.Duration
	TagFailures     bool
	ScoreStability  bool
}

type Result struct {
	Config         *common.NormalizedConfig
	Reachable      bool
	Attempts       int
	Successes      int
	AvgMs          int64
	MinMs          int64
	MaxMs          int64
	StabilityScore float64
	FailureTag     string
}

var (
	DefaultTimeout         = 5 * time.Second
	DefaultStabilityWindow = 0 * time.Second
)

type Tester struct {
	config  Config
	limiter chan struct{}
}

func IsSuspiciouslyFast(ms int64) bool {
	return ms < 10
}

func MostCommon(tags []string) string {
	counts := make(map[string]int)
	for _, t := range tags {
		counts[t]++
	}

	max := 0
	common := ""
	for tag, count := range counts {
		if count > max {
			max = count
			common = tag
		}
	}
	return common
}
