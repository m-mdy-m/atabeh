package command

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/internal/normalizer"
	"github.com/m-mdy-m/atabeh/internal/parsers"
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) SubCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sub",
		Short: "Manage subscription URLs",
		Long: `Subscription management - save URLs, sync them, and test configs.

Examples:
  # Save a subscription
  atabeh sub add https://example.com/sub

  # List saved subscriptions  
  atabeh sub list

  # Sync all subscriptions with concurrent testing
  atabeh sub sync-all --test-first --concurrent 30

  # Remove a subscription
  atabeh sub remove https://example.com/sub`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		c.subAddCmd(),
		c.subListCmd(),
		c.subRemoveCmd(),
		c.subSyncAllCmd(),
		c.subSyncOneCmd(),
	)
	return cmd
}

func (c *CLI) subAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <url>",
		Short: "Save a subscription URL",
		Long: `Save a subscription URL for later syncing.
The URL is stored but configs are NOT fetched yet.

Use 'sub sync-all' or 'sub sync <url>' to fetch configs.

Examples:
  atabeh sub add https://example.com/sub
  atabeh sub add https://raw.githubusercontent.com/user/repo/main/configs.txt`,
		Args: cobra.ExactArgs(1),
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			url := args[0]

			// Validate URL format
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				return fmt.Errorf("invalid URL: must start with http:// or https://")
			}

			// Add to subscriptions table
			err := repo.AddSubscription(url)
			if err != nil {
				return err
			}

			fmt.Printf("  ✓ Saved subscription: %s\n", url)
			fmt.Printf("\n  Use 'atabeh sub sync %s' to fetch configs\n", truncateURL(url, 40))
			return nil
		}),
	}
}

func (c *CLI) subListCmd() *cobra.Command {
	var showStats bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show all saved subscriptions",
		Long: `Lists all subscription URLs and their associated profiles.

Examples:
  atabeh sub list
  atabeh sub list --stats`,
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
				fmt.Printf("  %d. %s\n", i+1, url)

				if showStats {
					// Find profile for this subscription
					profiles, err := repo.ListProfiles()
					if err == nil {
						for _, p := range profiles {
							if p.Source == url || p.Source == "subscription:"+url {
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

	cmd.Flags().BoolVar(&showStats, "stats", false, "show profile statistics for each subscription")
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
		Long: `Removes a subscription URL and optionally its profile and configs.

Examples:
  # Remove subscription only (keep profile and configs)
  atabeh sub remove https://example.com/sub

  # Remove all subscriptions
  atabeh sub remove --all

  # Skip confirmation
  atabeh sub remove https://example.com/sub --yes`,
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			if removeAll {
				return removeAllSubscriptions(repo, confirm)
			}

			if len(args) != 1 {
				return fmt.Errorf("provide a URL or use --all")
			}

			url := args[0]
			return removeSingleSubscription(repo, url, confirm)
		}),
	}

	cmd.Flags().BoolVar(&removeAll, "all", false, "remove all saved subscriptions")
	cmd.Flags().BoolVarP(&confirm, "yes", "y", false, "skip confirmation prompt")
	return cmd
}

func (c *CLI) subSyncOneCmd() *cobra.Command {
	var (
		testFirst  bool
		runTest    bool
		concurrent int
	)

	cmd := &cobra.Command{
		Use:   "sync <url>",
		Short: "Sync a specific subscription",
		Long: `Fetch and sync configs from a specific subscription URL.

Examples:
  atabeh sub sync https://example.com/sub
  atabeh sub sync https://example.com/sub --test-first --concurrent 30`,
		Args: cobra.ExactArgs(1),
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			url := args[0]
			return syncSubscription(repo, url, testFirst, runTest, concurrent)
		}),
	}

	cmd.Flags().BoolVar(&testFirst, "test-first", false, "test configs before saving")
	cmd.Flags().BoolVar(&runTest, "test", false, "test configs after saving")
	cmd.Flags().IntVar(&concurrent, "concurrent", 20, "number of concurrent tests")
	return cmd
}

func (c *CLI) subSyncAllCmd() *cobra.Command {
	var (
		testFirst  bool
		runTest    bool
		concurrent int
	)

	cmd := &cobra.Command{
		Use:   "sync-all",
		Short: "Sync all saved subscriptions",
		Long: `Fetch and sync configs from every saved subscription.

Examples:
  atabeh sub sync-all
  atabeh sub sync-all --test-first --concurrent 30`,
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
			failedSubs := []string{}

			for i, url := range subs {
				fmt.Printf("  [%d/%d] %s\n", i+1, len(subs), truncateURL(url, 60))

				err := syncSubscription(repo, url, testFirst, false, concurrent)
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

			// Test all if requested
			if runTest {
				fmt.Println()
				return testAllProfiles(repo, concurrent)
			}

			return nil
		}),
	}

	cmd.Flags().BoolVar(&testFirst, "test-first", false, "test configs before saving")
	cmd.Flags().BoolVar(&runTest, "test", false, "test all configs after syncing")
	cmd.Flags().IntVar(&concurrent, "concurrent", 20, "number of concurrent tests")
	return cmd
}

// ========== Helper Functions ==========

func syncSubscription(repo *repository.Repo, url string, testFirst, runTest bool, concurrent int) error {
	// Fetch and parse
	logger.Infof("sub", "fetching: %s", url)
	rawConfigs, err := parsers.FetchAndParseAll(url)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	fetched := len(rawConfigs)
	logger.Infof("sub", "fetched %d configs", fetched)

	// Normalize
	configs, err := normalizer.Normalize(rawConfigs)
	if err != nil {
		return fmt.Errorf("normalize: %w", err)
	}
	logger.Infof("sub", "normalized %d configs", len(configs))

	if len(configs) == 0 {
		return fmt.Errorf("no valid configs found")
	}

	// Test first if requested
	if testFirst {
		logger.Infof("sub", "testing configs before save...")
		configs = testAndFilterConfigs(configs, concurrent)
		if len(configs) == 0 {
			return fmt.Errorf("no configs passed tests")
		}
		logger.Infof("sub", "%d configs passed tests", len(configs))
	}

	// Get or create profile
	profileName := parsers.ExtractProfileName(url)
	profileID, err := repo.GetOrCreateProfile(profileName, url, "subscription")
	if err != nil {
		return fmt.Errorf("create profile: %w", err)
	}

	// Insert configs
	inserted, err := repo.InsertConfigBatch(configs, profileID)
	if err != nil {
		return fmt.Errorf("insert configs: %w", err)
	}

	// Update sync time
	if err := repo.UpdateProfileSyncTime(profileID); err != nil {
		logger.Warnf("sub", "update sync time: %v", err)
	}

	// Get totals
	total, _ := repo.CountConfigsByProfile(int(profileID))

	fmt.Printf("       Profile:  %s\n", profileName)
	fmt.Printf("       Fetched:  %d\n", fetched)
	fmt.Printf("       Added:    %d new\n", inserted)
	fmt.Printf("       Skipped:  %d duplicate(s)\n", len(configs)-inserted)
	fmt.Printf("       Total:    %d in profile\n", total)

	// Test after save if requested
	if runTest && !testFirst {
		logger.Infof("sub", "testing saved configs...")
		return testProfileConfigs(repo, int(profileID), concurrent)
	}

	return nil
}

func removeSingleSubscription(repo *repository.Repo, url string, skipConfirm bool) error {
	// Check if subscription exists
	subs, err := repo.ListSubscriptions()
	if err != nil {
		return err
	}

	found := false
	for _, s := range subs {
		if s == url {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("subscription not found: %s", url)
	}

	// Find associated profile
	var profile *storage.ProfileRow
	profiles, err := repo.ListProfiles()
	if err == nil {
		for _, p := range profiles {
			if p.Source == url || p.Source == "subscription:"+url {
				profile = p
				break
			}
		}
	}

	// Confirmation
	if !skipConfirm {
		fmt.Printf("  Remove subscription:\n")
		fmt.Printf("    URL: %s\n", url)
		if profile != nil {
			fmt.Printf("    Profile: %s (%d configs)\n", profile.Name, profile.ConfigCount)
			fmt.Printf("\n  Note: Profile and configs will remain. Use 'atabeh profile delete %d' to remove them.\n", profile.ID)
		}
		fmt.Printf("\n  Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) != "y" {
			fmt.Println("  Cancelled.")
			return nil
		}
	}

	// Remove subscription
	if err := repo.RemoveSubscription(url); err != nil {
		return err
	}

	fmt.Printf("  ✓ Removed subscription\n")
	if profile != nil {
		fmt.Printf("  Profile '%s' and its configs are still available\n", profile.Name)
	}

	return nil
}

func removeAllSubscriptions(repo *repository.Repo, skipConfirm bool) error {
	subs, err := repo.ListSubscriptions()
	if err != nil {
		return err
	}

	if len(subs) == 0 {
		fmt.Println("  No subscriptions to remove.")
		return nil
	}

	// Confirmation
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

	// Remove all
	if err := repo.ClearSubscriptions(); err != nil {
		return err
	}

	fmt.Printf("  ✓ Removed %d subscription(s)\n", len(subs))
	return nil
}

func testAllProfiles(repo *repository.Repo, concurrent int) error {
	profiles, err := repo.ListProfiles()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		fmt.Println("  No profiles to test.")
		return nil
	}

	fmt.Printf("  Testing all profiles...\n\n")

	for _, profile := range profiles {
		if profile.ConfigCount == 0 {
			continue
		}

		fmt.Printf("  Profile: %s\n", profile.Name)
		if err := testProfileConfigs(repo, profile.ID, concurrent); err != nil {
			logger.Warnf("test", "profile %d: %v", profile.ID, err)
		}
		fmt.Println()
	}

	return nil
}

func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}
