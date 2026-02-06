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

type TestConfig struct {
	Attempts        int
	TimeoutSec      int
	Concurrent      int
	DelaySec        int
	StabilityWindow int
	TagReasons      bool
	StabilityScore  bool
}

func (c *CLI) TestCommand() *cobra.Command {
	var (
		testAll     bool
		testID      int
		testProfile int
		cfg         TestConfig
	)

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test config connectivity with advanced validation",
		Long: `Advanced connectivity testing with:
  - Fake ping detection
  - Stability scoring over time
  - Detailed failure tagging (DPI, TLS, DNS, etc.)
  - Concurrent testing for speed

Examples:
  atabeh test --all
  atabeh test --id 5
  atabeh test --profile 2
  atabeh test --all --tag-reasons
  atabeh test --all --stability-window 60 --stability-score
  atabeh test --all --attempts 7 --delay 5`,
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			testerCfg := tester.Config{
				Attempts:        cfg.Attempts,
				Timeout:         time.Duration(cfg.TimeoutSec) * time.Second,
				ConcurrentTests: cfg.Concurrent,
				TestDelay:       time.Duration(cfg.DelaySec) * time.Second,
				StabilityWindow: time.Duration(cfg.StabilityWindow) * time.Second,
				TagFailures:     cfg.TagReasons,
				ScoreStability:  cfg.StabilityScore,
			}

			switch {
			case testID > 0:
				return runSingleTest(repo, testerCfg, testID)
			case testProfile > 0:
				return runProfileTests(repo, testerCfg, testProfile)
			case testAll:
				return runAllTests(repo, testerCfg)
			default:
				return fmt.Errorf("specify --all, --profile <ID>, or --id <N>")
			}
		}),
	}

	cmd.Flags().BoolVar(&testAll, "all", false, "test all configs")
	cmd.Flags().IntVar(&testID, "id", 0, "test single config")
	cmd.Flags().IntVar(&testProfile, "profile", 0, "test profile configs")
	cmd.Flags().IntVar(&cfg.Attempts, "attempts", 3, "attempts per config")
	cmd.Flags().IntVar(&cfg.TimeoutSec, "timeout", 5, "connection timeout (sec)")
	cmd.Flags().IntVar(&cfg.Concurrent, "concurrent", 20, "concurrent tests")
	cmd.Flags().IntVar(&cfg.DelaySec, "delay", 0, "delay between attempts (sec)")
	cmd.Flags().IntVar(&cfg.StabilityWindow, "stability-window", 0, "test stability over time (sec)")
	cmd.Flags().BoolVar(&cfg.TagReasons, "tag-reasons", false, "tag failure reasons (DPI, TLS, DNS)")
	cmd.Flags().BoolVar(&cfg.StabilityScore, "stability-score", false, "score by stability not latency")

	return cmd
}

func runSingleTest(repo *repository.Repo, cfg tester.Config, id int) error {
	stored, err := repo.GetConfigByID(id)
	if err != nil {
		return fmt.Errorf("config %d: %w", id, err)
	}

	norm := toNormalized(stored)
	result := tester.Test(norm, cfg)

	if err := repo.UpdateConfigPingResult(id, result); err != nil {
		return err
	}

	printTestResult(stored, result, cfg.TagFailures)
	return nil
}

func runProfileTests(repo *repository.Repo, cfg tester.Config, profileID int) error {
	profile, err := repo.GetProfile(profileID)
	if err != nil {
		return err
	}

	fmt.Printf("\n  Testing profile: %s\n\n", profile.Name)

	configs, err := repo.ListConfigsByProfile(profileID)
	if err != nil {
		return err
	}

	if len(configs) == 0 {
		fmt.Println("  No configs in profile")
		return nil
	}

	return runTestBatch(repo, cfg, configs)
}

func runAllTests(repo *repository.Repo, cfg tester.Config) error {
	configs, err := repo.ListConfigs("")
	if err != nil {
		return err
	}

	if len(configs) == 0 {
		fmt.Println("  No configs. Use `atabeh add` first.")
		return nil
	}

	fmt.Printf("\n  Testing %d configs...\n\n", len(configs))
	return runTestBatch(repo, cfg, configs)
}

func runTestBatch(repo *repository.Repo, cfg tester.Config, storeds []*storage.ConfigRow) error {
	norms := make([]*common.NormalizedConfig, len(storeds))
	for i, s := range storeds {
		norms[i] = toNormalized(s)
	}

	logger.Infof("test", "testing %d configs (concurrent=%d)", len(norms), cfg.ConcurrentTests)

	results := tester.TestAll(norms, cfg)

	// Save results
	resultMap := make(map[int]*tester.Result)
	for i, r := range results {
		resultMap[storeds[i].ID] = r
	}

	if err := repo.UpdateConfigPingBatch(resultMap); err != nil {
		logger.Errorf("test", "save results: %v", err)
	}

	printTestSummary(storeds, results, cfg.TagFailures, cfg.ScoreStability)
	return nil
}
