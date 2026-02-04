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
	var (
		testAll     bool
		testID      int
		testProfile int
		testTimeout int
		concurrent  int
	)

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test connectivity of configs",
		Long: `Runs concurrent TCP connectivity tests against stored configs.

Examples:
  atabeh test --all
  atabeh test --id 3
  atabeh test --profile 1
  atabeh test --all --concurrent 30 --timeout 10`,
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			cfg := tester.Config{
				Attempts:        3,
				Timeout:         time.Duration(testTimeout) * time.Second,
				ConcurrentTests: concurrent,
				TestDelay:       100 * time.Millisecond,
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
	cmd.Flags().IntVar(&testTimeout, "timeout", 5, "per-connection timeout in seconds")
	cmd.Flags().IntVar(&concurrent, "concurrent", 20, "number of concurrent tests")
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

	// Test all concurrently
	results := tester.TestAll(norms, cfg)

	// Update database in batch
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
