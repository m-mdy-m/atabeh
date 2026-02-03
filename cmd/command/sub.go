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

func (c *CLI) SubCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sub",
		Short: "Manage subscription URLs",
		Long: `Subscription management.  Save URLs once, sync them any time.

  atabeh sub add    https://raw.githubusercontent.com/…
  atabeh sub list
  atabeh sub remove https://…
  atabeh sub sync-all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		c.subAddCmd(),
		c.subListCmd(),
		c.subRemoveCmd(),
		c.subSyncAllCmd(),
	)
	return cmd
}

func (c *CLI) subAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <url>",
		Short: "Save a subscription URL",
		Long: `Save a subscription URL so you can sync it later with ` + "`sub sync-all`" + `.
The URL is stored in the database; configs are NOT fetched yet.

Examples:
  atabeh sub add https://raw.githubusercontent.com/user/repo/main/sub
  atabeh sub add https://sub.example.com/configs`,
		Args: cobra.ExactArgs(1),
		RunE: c.WrapRepo(func(repo *storage.ConfigRepo, cmd *cobra.Command, args []string) error {
			url := args[0]
			err := repo.AddSubscription(url)
			if err != nil {
				return err
			}
			fmt.Printf("  saved subscription: %s\n", url)
			return nil
		}),
	}
}

func (c *CLI) subListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show all saved subscriptions",
		RunE: c.WrapRepo(func(repo *storage.ConfigRepo, cmd *cobra.Command, args []string) error {
			subs, err := repo.ListSubscriptions()
			if err != nil {
				return err
			}
			if len(subs) == 0 {
				fmt.Println("  No subscriptions saved. Use `atabeh sub add <url>`.")
				return nil
			}
			for i, s := range subs {
				fmt.Printf("  %d.  %s\n", i+1, s)
			}
			return nil
		}),
	}
}

func (c *CLI) subRemoveCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "remove [url]",
		Short: "Remove a saved subscription (and its configs)",
		Long: `Removes a subscription URL and all configs that came from it.

Examples:
  atabeh sub remove https://sub.example.com/configs
  atabeh sub remove --all`,
		RunE: c.WrapRepo(func(repo *storage.ConfigRepo, cmd *cobra.Command, args []string) error {
			if all {
				if err := repo.ClearSubscriptions(); err != nil {
					return err
				}
				fmt.Println("  all subscriptions removed")
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("provide a URL or --all")
			}
			url := args[0]

			// remove configs sourced from this sub
			n, err := repo.DeleteBySource("subscription:" + url)
			if err != nil {
				return err
			}
			// remove the subscription record itself
			if err := repo.RemoveSubscription(url); err != nil {
				return err
			}
			fmt.Printf("  removed subscription + %d config(s)\n", n)
			return nil
		}),
	}
	cmd.Flags().BoolVar(&all, "all", false, "remove every saved subscription")
	return cmd
}

func (c *CLI) subSyncAllCmd() *cobra.Command {
	var runTest bool

	cmd := &cobra.Command{
		Use:   "sync-all",
		Short: "Fetch & store configs from every saved subscription",
		Long: `Iterates over every URL saved with ` + "`sub add`" + ` and syncs each one.

  atabeh sub sync-all
  atabeh sub sync-all --test`,
		RunE: c.WrapRepo(func(repo *storage.ConfigRepo, cmd *cobra.Command, args []string) error {
			subs, err := repo.ListSubscriptions()
			if err != nil {
				return err
			}
			if len(subs) == 0 {
				fmt.Println("  No subscriptions saved. Use `atabeh sub add <url>` first.")
				return nil
			}

			totalFetched, totalInserted := 0, 0

			for _, url := range subs {
				fmt.Printf("\n  syncing %s\n", url)

				raw, err := parsers.FetchSubscription(url)
				if err != nil {
					logger.Warnf("sub", "fetch %s: %v", url, err)
					continue
				}
				totalFetched += len(raw)

				configs, err := normalizer.Normalize(raw)
				if err != nil {
					logger.Warnf("sub", "normalize %s: %v", url, err)
					continue
				}

				for _, cfg := range configs {
					_, isNew, err := repo.InsertOrSkip(cfg, "subscription:"+url)
					if err != nil {
						logger.Warnf("sub", "insert %q: %v", cfg.Name, err)
						continue
					}
					if isNew {
						totalInserted++
					}
				}
			}

			total, _ := repo.Count()
			fmt.Println()
			printSyncSummary("(all subscriptions)", totalFetched, totalInserted, total)

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
