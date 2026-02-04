package command

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/internal/normalizer"
	"github.com/m-mdy-m/atabeh/internal/parsers"
	"github.com/m-mdy-m/atabeh/internal/tester"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) SyncCommand() *cobra.Command {
	var (
		runTest    bool
		testFirst  bool
		concurrent int
	)

	cmd := &cobra.Command{
		Use:   "sync <source>",
		Short: "Fetch, test and store configs from a source",
		Long: `Downloads configs from source (subscription URL or mixed content text),
validates, tests (optionally), and stores valid configs in a profile.

Source can be:
  - Subscription URL
  - Text file path
  - Raw text with configs

Examples:
  atabeh sync https://example.com/sub
  atabeh sync https://example.com/sub --test
  atabeh sync https://example.com/sub --test-first --concurrent 20`,
		Args: cobra.MinimumNArgs(1),
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			source := args[0]

			logger.Infof("sync", "fetching from: %s", source)
			rawConfigs, err := parsers.FetchAndParseAll(source)
			if err != nil {
				return fmt.Errorf("fetch: %w", err)
			}
			fetched := len(rawConfigs)
			logger.Infof("sync", "fetched %d raw configs", fetched)

			configs, err := normalizer.Normalize(rawConfigs)
			if err != nil {
				return fmt.Errorf("normalize: %w", err)
			}
			logger.Infof("sync", "normalized %d configs", len(configs))

			if testFirst {
				logger.Infof("sync", "testing configs before saving...")
				configs = testAndFilterConfigs(configs, concurrent)
				logger.Infof("sync", "%d configs passed tests", len(configs))
			}

			profileName := parsers.ExtractProfileName(source)
			profileType := "subscription"
			if !isURL(source) {
				profileType = "mixed"
			}

			profileID, err := repo.GetOrCreateProfile(profileName, source, profileType)
			if err != nil {
				return fmt.Errorf("create profile: %w", err)
			}

			inserted, err := repo.InsertConfigBatch(configs, profileID)
			if err != nil {
				return fmt.Errorf("insert configs: %w", err)
			}

			if err := repo.UpdateProfileSyncTime(profileID); err != nil {
				logger.Warnf("sync", "update sync time: %v", err)
			}

			total, _ := repo.CountConfigsByProfile(int(profileID))
			printSyncSummary(profileName, fetched, inserted, total)

			if runTest && !testFirst {
				logger.Infof("sync", "testing saved configs...")
				return testProfileConfigs(repo, int(profileID), concurrent)
			}

			return nil
		}),
	}

	cmd.Flags().BoolVar(&runTest, "test", false, "test configs after syncing")
	cmd.Flags().BoolVar(&testFirst, "test-first", false, "test configs before saving (only save working ones)")
	cmd.Flags().IntVar(&concurrent, "concurrent", 20, "number of concurrent tests")
	return cmd
}

func testAndFilterConfigs(configs []*common.NormalizedConfig, concurrent int) []*common.NormalizedConfig {
	if len(configs) == 0 {
		return configs
	}

	tcfg := tester.Config{
		Attempts:        2,
		Timeout:         5 * time.Second,
		ConcurrentTests: concurrent,
		TestDelay:       50 * time.Millisecond,
	}

	results := tester.TestAll(configs, tcfg)

	working := make([]*common.NormalizedConfig, 0)
	for i, result := range results {
		if result.Reachable {
			working = append(working, configs[i])
		}
	}

	return working
}

func testProfileConfigs(repo *repository.Repo, profileID int, concurrent int) error {
	storeds, err := repo.ListConfigsByProfile(profileID)
	if err != nil {
		return err
	}

	if len(storeds) == 0 {
		fmt.Println("  No configs to test.")
		return nil
	}

	norms := make([]*common.NormalizedConfig, len(storeds))
	for i, s := range storeds {
		norms[i] = toNormalized(s)
	}

	tcfg := tester.Config{
		Attempts:        3,
		Timeout:         5 * time.Second,
		ConcurrentTests: concurrent,
		TestDelay:       100 * time.Millisecond,
	}

	logger.Infof("test", "testing %d configs with concurrency=%d", len(norms), concurrent)
	results := tester.TestAll(norms, tcfg)

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

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
