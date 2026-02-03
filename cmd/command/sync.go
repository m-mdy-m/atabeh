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

type syncOptions struct {
	test bool
}

func (c *CLI) SyncCommand() *cobra.Command {
	opts := &syncOptions{}

	cmd := &cobra.Command{
		Use:   "sync <subscription-url>",
		Short: "Fetch, parse and store configs from a subscription URL",
		Long: `Downloads the config list from the given subscription URL (base64 or
plain-text URI list), parses every entry, validates & deduplicates,
then stores the new ones in the local database.

Examples:
  atabeh sync https://sub.example.com/configs
  atabeh sync https://sub.example.com/configs --test`,
		Args: cobra.ExactArgs(1),

		RunE: c.WrapRepo(func(repo *storage.ConfigRepo, cmd *cobra.Command, args []string) error {
			url := args[0]

			raw, err := parsers.FetchSubscription(url)
			if err != nil {
				return fmt.Errorf("fetch/parse subscription: %w", err)
			}
			fetched := len(raw)
			logger.Infof("sync", "fetched %d raw config(s)", fetched)
			configs, err := normalizer.Normalize(raw)
			if err != nil {
				return fmt.Errorf("normalisation: %w", err)
			}
			inserted := 0
			for _, cfg := range configs {
				_, isNew, err := repo.InsertOrSkip(cfg, "subscription:"+url)
				if err != nil {
					logger.Warnf("sync", "insert failed for %q: %v", cfg.Name, err)
					continue
				}
				if isNew {
					inserted++
				}
			}

			total, _ := repo.Count()
			logger.StorageReport("sync", url, fetched, inserted, total)
			if opts.test {
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

	cmd.Flags().BoolVar(&opts.test, "test", false, "run connectivity tests after syncing")
	return cmd
}
