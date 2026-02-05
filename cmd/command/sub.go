package command

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/internal/normalizer"
	"github.com/m-mdy-m/atabeh/internal/parsers"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) SubCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sub",
		Short: "Manage subscription URLs",
		Long: `Subscription management - save URLs, sync them, test configs.

Examples:
  atabeh sub add <url>               # Save subscription
  atabeh sub list                    # List saved subscriptions  
  atabeh sub sync                    # Sync latest subscription
  atabeh sub sync <url>              # Sync specific subscription
  atabeh sub sync-all                # Sync all subscriptions
  atabeh sub remove <url>            # Remove subscription`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		c.subAddCmd(),
		c.subListCmd(),
		c.subSyncCmd(),
		c.subSyncAllCmd(),
		c.subRemoveCmd(),
	)
	return cmd
}

func (c *CLI) subAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <url>",
		Short: "Save a subscription URL",
		Args:  cobra.ExactArgs(1),
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			url := args[0]

			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				return fmt.Errorf("invalid URL: must start with http:// or https://")
			}

			err := repo.AddSubscription(url)
			if err != nil {
				return err
			}

			fmt.Printf("  ✓ Saved subscription: %s\n", truncateURL(url, 60))
			fmt.Printf("\n  Use 'atabeh sub sync' to fetch configs\n")
			return nil
		}),
	}
}

func (c *CLI) subListCmd() *cobra.Command {
	var showStats bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show all saved subscriptions",
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			subs, err := repo.ListSubscriptions()
			if err != nil {
				return err
			}

			if len(subs) == 0 {
				fmt.Println("  No subscriptions saved.")
				fmt.Println("  Use `atabeh sub add <url>` to add one.")
				return nil
			}

			fmt.Printf("  Saved subscriptions (%d):\n\n", len(subs))

			for i, url := range subs {
				fmt.Printf("  %d. %s\n", i+1, truncateURL(url, 70))

				if showStats {
					profiles, err := repo.ListProfiles()
					if err == nil {
						for _, p := range profiles {
							if p.Source == url {
								fmt.Printf("     → Profile: %s (%d configs, %d alive)\n",
									p.Name, p.ConfigCount, p.AliveCount)
								if p.LastSyncedAt != "" {
									fmt.Printf("     → Last sync: %s\n", p.LastSyncedAt)
								}
								break
							}
						}
					}
					fmt.Println()
				}
			}

			if !showStats {
				fmt.Printf("\n  Use --stats to see profile information\n")
			}

			return nil
		}),
	}

	cmd.Flags().BoolVar(&showStats, "stats", false, "show profile stats")
	return cmd
}

func (c *CLI) subSyncCmd() *cobra.Command {
	var (
		testFirst  bool
		concurrent int
	)

	cmd := &cobra.Command{
		Use:   "sync [url]",
		Short: "Sync subscription (latest if no URL given)",
		Long: `Fetch and sync configs from subscription.
If no URL provided, syncs the latest added subscription.

Examples:
  atabeh sub sync                     # Sync latest subscription
  atabeh sub sync <url>               # Sync specific URL
  atabeh sub sync --test-first        # Test before saving`,
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			var url string
			var err error

			if len(args) == 0 {
				url, err = repo.GetLatestSubscription()
				if err != nil {
					return fmt.Errorf("no subscriptions found. Use 'atabeh sub add <url>' first")
				}
				fmt.Printf("  Using latest subscription: %s\n\n", truncateURL(url, 60))
			} else {
				url = args[0]
				exists, err := repo.SubscriptionExists(url)
				if err != nil {
					return err
				}
				if !exists {
					return fmt.Errorf("subscription not found: %s\nUse 'atabeh sub add <url>' first", url)
				}
			}

			return syncSubscription(repo, url, testFirst, concurrent)
		}),
	}

	cmd.Flags().BoolVar(&testFirst, "test-first", false, "test configs before saving")
	cmd.Flags().IntVar(&concurrent, "concurrent", 20, "concurrent tests")
	return cmd
}

func (c *CLI) subSyncAllCmd() *cobra.Command {
	var (
		testFirst  bool
		concurrent int
	)

	cmd := &cobra.Command{
		Use:   "sync-all",
		Short: "Sync all saved subscriptions",
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			subs, err := repo.ListSubscriptions()
			if err != nil {
				return err
			}

			if len(subs) == 0 {
				fmt.Println("  No subscriptions saved.")
				fmt.Println("  Use `atabeh sub add <url>` first.")
				return nil
			}

			fmt.Printf("  Syncing %d subscription(s)...\n\n", len(subs))

			successCount := 0
			var failedSubs []string

			for i, url := range subs {
				fmt.Printf("  [%d/%d] %s\n", i+1, len(subs), truncateURL(url, 60))

				err := syncSubscription(repo, url, testFirst, concurrent)
				if err != nil {
					logger.Warnf("sub", "sync %s: %v", url, err)
					failedSubs = append(failedSubs, url)
					fmt.Printf("       ✗ Failed: %v\n\n", err)
					continue
				}

				successCount++
				fmt.Println()
			}

			// Summary
			fmt.Printf("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("  Sync complete:\n")
			fmt.Printf("    Success: %d/%d subscriptions\n", successCount, len(subs))
			if len(failedSubs) > 0 {
				fmt.Printf("    Failed:  %d\n", len(failedSubs))
				for _, url := range failedSubs {
					fmt.Printf("      - %s\n", truncateURL(url, 60))
				}
			}

			return nil
		}),
	}

	cmd.Flags().BoolVar(&testFirst, "test-first", false, "test before saving")
	cmd.Flags().IntVar(&concurrent, "concurrent", 20, "concurrent tests")
	return cmd
}

func (c *CLI) subRemoveCmd() *cobra.Command {
	var (
		removeAll bool
		confirm   bool
	)

	cmd := &cobra.Command{
		Use:   "remove [url]",
		Short: "Remove a saved subscription",
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			if removeAll {
				return handleRemoveAllSubs(repo, confirm)
			}

			if len(args) != 1 {
				return fmt.Errorf("provide a URL or use --all")
			}

			return handleRemoveSub(repo, args[0], confirm)
		}),
	}

	cmd.Flags().BoolVar(&removeAll, "all", false, "remove all subscriptions")
	cmd.Flags().BoolVarP(&confirm, "yes", "y", false, "skip confirmation")
	return cmd
}

func handleRemoveSub(repo *repository.Repo, url string, skipConfirm bool) error {
	exists, err := repo.SubscriptionExists(url)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("subscription not found: %s", url)
	}

	if !skipConfirm {
		fmt.Printf("  Remove subscription: %s\n", truncateURL(url, 60))
		fmt.Printf("  Note: Profile and configs will remain.\n")
		fmt.Printf("\n  Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("  Cancelled.")
			return nil
		}
	}

	if err := repo.RemoveSubscription(url); err != nil {
		return err
	}

	fmt.Printf("  ✓ Removed subscription\n")
	return nil
}

func handleRemoveAllSubs(repo *repository.Repo, skipConfirm bool) error {
	subs, err := repo.ListSubscriptions()
	if err != nil {
		return err
	}

	if len(subs) == 0 {
		fmt.Println("  No subscriptions to remove.")
		return nil
	}

	if !skipConfirm {
		fmt.Printf("  Remove all %d subscription(s)?\n", len(subs))
		fmt.Printf("  Note: Profiles and configs will remain.\n")
		fmt.Printf("\n  Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("  Cancelled.")
			return nil
		}
	}

	if err := repo.ClearSubscriptions(); err != nil {
		return err
	}

	fmt.Printf("  ✓ Removed %d subscription(s)\n", len(subs))
	return nil
}

func syncSubscription(repo *repository.Repo, url string, testFirst bool, concurrent int) error {
	logger.Infof("sub", "fetching: %s", url)

	rawConfigs, err := parsers.FetchAndParseAll(url)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	fetched := len(rawConfigs)
	logger.Infof("sub", "fetched %d configs", fetched)

	configs, err := normalizer.Normalize(rawConfigs)
	if err != nil {
		return fmt.Errorf("normalize: %w", err)
	}
	logger.Infof("sub", "normalized %d configs", len(configs))

	if len(configs) == 0 {
		return fmt.Errorf("no valid configs found")
	}

	if testFirst {
		logger.Infof("sub", "testing configs before save...")
		configs = testAndFilterConfigs(configs, concurrent)
		if len(configs) == 0 {
			return fmt.Errorf("no configs passed tests")
		}
		logger.Infof("sub", "%d configs passed tests", len(configs))
	}
	profileName := parsers.ExtractProfileName(url)
	profileID, err := repo.GetOrCreateProfile(profileName, url, "subscription")
	if err != nil {
		return fmt.Errorf("create profile: %w", err)
	}
	inserted, err := repo.InsertConfigBatch(configs, profileID)
	if err != nil {
		return fmt.Errorf("insert configs: %w", err)
	}
	if err := repo.UpdateProfileSyncTime(profileID); err != nil {
		logger.Warnf("sub", "update sync time: %v", err)
	}
	total, _ := repo.CountConfigsByProfile(int(profileID))

	fmt.Printf("       Profile:  %s\n", profileName)
	fmt.Printf("       Fetched:  %d\n", fetched)
	fmt.Printf("       Added:    %d new\n", inserted)
	fmt.Printf("       Skipped:  %d duplicate(s)\n", len(configs)-inserted)
	fmt.Printf("       Total:    %d in profile\n", total)

	return nil
}

func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}
