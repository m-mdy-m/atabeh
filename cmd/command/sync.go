package command

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/internal/normalizer"
	"github.com/m-mdy-m/atabeh/internal/parsers"
	"github.com/m-mdy-m/atabeh/internal/storage"
	"github.com/m-mdy-m/atabeh/internal/tester"
)

func (c *CLI) SyncCommand() *cobra.Command {
	var runTest bool

	cmd := &cobra.Command{
		Use:   "sync <subscription-url>",
		Short: "Fetch, parse and store configs from a subscription URL",
		Long: `Downloads configs from a subscription URL (base64 or plain-text),
parses every entry, validates, deduplicates, and stores new ones.

Examples:
  atabeh sync https://sub.example.com/configs
  atabeh sync https://raw.githubusercontent.com/user/repo/main/sub --test`,
		Args: cobra.ExactArgs(1),
		RunE: c.WrapRepo(func(repo *storage.ConfigRepo, cmd *cobra.Command, args []string) error {
			url := args[0]

			raw, err := parsers.FetchSubscription(url)
			if err != nil {
				return fmt.Errorf("fetch: %w", err)
			}
			fetched := len(raw)

			configs, err := normalizer.Normalize(raw)
			if err != nil {
				return fmt.Errorf("normalize: %w", err)
			}

			inserted := 0
			for _, cfg := range configs {
				_, isNew, err := repo.InsertOrSkip(cfg, "subscription:"+url)
				if err != nil {
					logger.Warnf("sync", "insert %q: %v", cfg.Name, err)
					continue
				}
				if isNew {
					inserted++
				}
			}

			total, _ := repo.Count()
			printSyncSummary(url, fetched, inserted, total)

			if runTest {
				tcfg := tester.Config{
					Attempts:        3,
					Timeout:         5 * time.Second,
					ConcurrentTests: 10,
					TestDelay:       100 * time.Millisecond,
				}
				return runAll(repo, tcfg)
			}
			return nil
		}),
	}

	cmd.Flags().BoolVar(&runTest, "test", false, "run connectivity tests after syncing")
	return cmd
}
