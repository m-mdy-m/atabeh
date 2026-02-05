package command

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/internal/tester"
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) TestCommand() *cobra.Command {
	baseCfg := tester.DefaultConfig()

	var (
		testAll          bool
		testID           int
		testProfile      int
		testTimeout      int
		concurrent       int
		attempts         = baseCfg.Attempts
		delayMs          = int(baseCfg.TestDelay / time.Millisecond)
		bandwidthFlag    = baseCfg.BandwidthTest
		bandwidthTimeout = int(baseCfg.BandwidthTimeout / time.Second)
		minBandwidthKBps = baseCfg.MinBandwidthKBps
	)

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test connectivity of configs",
		Long: `Runs concurrent TCP connectivity tests against stored configs.

Examples:
  atabeh test --all
  atabeh test --id 3
  atabeh test --profile 1
  atabeh test --all --concurrent 30 --timeout 10
  atabeh test --all --attempts 5 --delay 200 --bandwidth --bandwidth-timeout 8 --min-bandwidth 150`,
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			cfg := tester.Config{
				Attempts:         attempts,
				Timeout:          time.Duration(testTimeout) * time.Second,
				ConcurrentTests:  concurrent,
				TestDelay:        time.Duration(delayMs) * time.Millisecond,
				BandwidthTest:    bandwidthFlag,
				BandwidthTimeout: time.Duration(bandwidthTimeout) * time.Second,
				MinBandwidthKBps: minBandwidthKBps,
			}

			switch {
			case testID > 0:
				return runSingle(repo, cfg, testID)
			case testProfile > 0:
				return runProfileTest(repo, cfg, testProfile)
			case testAll:
				return runAllTest(repo, cfg)
			default:
				return fmt.Errorf("use --all, --profile <ID>, or --id <N>")
			}
		}),
	}

	cmd.Flags().BoolVar(&testAll, "all", false, "test all stored configs")
	cmd.Flags().IntVar(&testID, "id", 0, "test a single config by id")
	cmd.Flags().IntVar(&testProfile, "profile", 0, "test all configs in a profile")
	cmd.Flags().IntVar(&testTimeout, "timeout", int(baseCfg.Timeout/time.Second), "per-connection timeout in seconds")
	cmd.Flags().IntVar(&concurrent, "concurrent", baseCfg.ConcurrentTests, "number of concurrent tests")
	cmd.Flags().IntVar(&attempts, "attempts", attempts, "number of attempts per config")
	cmd.Flags().IntVar(&delayMs, "delay", delayMs, "delay between attempts in milliseconds")
	cmd.Flags().BoolVar(&bandwidthFlag, "bandwidth", bandwidthFlag, "enable bandwidth test to detect fake pings")
	cmd.Flags().IntVar(&bandwidthTimeout, "bandwidth-timeout", bandwidthTimeout, "timeout for bandwidth test in seconds")
	cmd.Flags().IntVar(&minBandwidthKBps, "min-bandwidth", minBandwidthKBps, "minimum acceptable bandwidth in KB/s (used when --bandwidth is set)")

	return cmd
}

func runProfileTest(repo *repository.Repo, cfg tester.Config, profileID int) error {
	profile, err := repo.GetProfile(profileID)
	if err != nil {
		return fmt.Errorf("get profile: %w", err)
	}

	fmt.Printf("\n  Testing profile: %s\n\n", profile.Name)

	storeds, err := repo.ListConfigsByProfile(profileID)
	if err != nil {
		return err
	}

	if len(storeds) == 0 {
		fmt.Println("  No configs in this profile.")
		return nil
	}

	return runTestsOnConfigs(repo, cfg, storeds)
}

func runAllTest(repo *repository.Repo, cfg tester.Config) error {
	storeds, err := repo.ListConfigs("")
	if err != nil {
		return err
	}

	if len(storeds) == 0 {
		fmt.Println("  Nothing to test. Add configs first.")
		return nil
	}

	fmt.Printf("\n  Testing all configs...\n\n")
	return runTestsOnConfigs(repo, cfg, storeds)
}

func runTestsOnConfigs(repo *repository.Repo, cfg tester.Config, storeds []*storage.ConfigRow) error {
	norms := make([]*common.NormalizedConfig, len(storeds))
	for i, s := range storeds {
		norms[i] = toNormalized(s)
	}

	logger.Infof("test", "testing %d configs with concurrency=%d",
		len(norms), cfg.ConcurrentTests)

	results := tester.TestAll(norms, cfg)

	resultMap := make(map[int]*common.PingResult)
	for i, result := range results {
		resultMap[storeds[i].ID] = result
	}

	if err := repo.UpdateConfigPingBatch(resultMap); err != nil {
		logger.Errorf("test", "update results: %v", err)
	}

	printTestSummary(storeds, results)
	return nil
}
