package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/internal/normalizer"
	"github.com/m-mdy-m/atabeh/internal/parsers"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) AddCommand() *cobra.Command {
	var (
		testFirst bool
		profile   string
	)

	cmd := &cobra.Command{
		Use:   "add <config_or_text>",
		Short: "Add config(s) from URI or mixed text",
		Long: `Parses config(s) from URI, text file, or mixed content and stores them.

Supports:
  - Single config URI
  - Multiple config URIs (space or newline separated)
  - Mixed content (configs + subscription URLs)
  - Text files with configs

Examples:
  atabeh add "vless://uuid@server:443?security=tls#MyServer"
  atabeh add "vless://... vmess://..." --test-first
  atabeh add @file_with_configs.txt --profile "My Configs"`,
		Args: cobra.MinimumNArgs(1),
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			input := args[0]

			// Read from file if starts with @
			if len(input) > 0 && input[0] == '@' {
				content, err := os.ReadFile(input[1:])
				if err != nil {
					return fmt.Errorf("read file: %w", err)
				}
				input = string(content)
			}

			// Parse mixed content
			mixed, err := parsers.ParseMixedContent(input)
			if err != nil {
				return fmt.Errorf("parse: %w", err)
			}

			allRaw := make([]*common.RawConfig, 0)
			allRaw = append(allRaw, mixed.DirectConfigs...)

			// Fetch from nested subscriptions
			for _, subURL := range mixed.Subscriptions {
				logger.Infof("add", "fetching subscription: %s", subURL)
				subConfigs, err := parsers.FetchAndParseAll(subURL)
				if err != nil {
					logger.Warnf("add", "fetch sub %s: %v", subURL, err)
					continue
				}
				allRaw = append(allRaw, subConfigs...)
			}

			if len(allRaw) == 0 {
				return fmt.Errorf("no valid configs found")
			}

			logger.Infof("add", "found %d raw configs", len(allRaw))

			// Normalize
			configs, err := normalizer.Normalize(allRaw)
			if err != nil {
				return fmt.Errorf("normalize: %w", err)
			}
			logger.Infof("add", "normalized to %d configs", len(configs))

			if len(configs) == 0 {
				return fmt.Errorf("all configs failed validation")
			}

			// Test first if requested
			if testFirst {
				logger.Infof("add", "testing configs...")
				configs = testAndFilterConfigs(configs, 20)
				if len(configs) == 0 {
					return fmt.Errorf("no configs passed tests")
				}
				logger.Infof("add", "%d configs passed tests", len(configs))
			}

			// Determine profile name
			profileName := profile
			if profileName == "" {
				if len(configs) == 1 {
					profileName = configs[0].Name
				} else {
					profileName = configs[len(configs)-1].Name
				}
			}

			// Create or get profile
			source := "manual"
			if len(mixed.Subscriptions) > 0 {
				source = mixed.Subscriptions[0]
			}

			profileType := "manual"
			if len(mixed.Subscriptions) > 0 && len(mixed.DirectConfigs) > 0 {
				profileType = "mixed"
			} else if len(mixed.Subscriptions) > 0 {
				profileType = "subscription"
			}

			profileID, err := repo.GetOrCreateProfile(profileName, source, profileType)
			if err != nil {
				return fmt.Errorf("create profile: %w", err)
			}

			// Insert configs
			inserted := 0
			for _, cfg := range configs {
				_, isNew, err := repo.InsertConfigOrSkip(cfg, profileID)
				if err != nil {
					logger.Warnf("add", "insert %q: %v", cfg.Name, err)
					continue
				}
				if isNew {
					inserted++
				}
			}

			total, _ := repo.CountConfigsByProfile(int(profileID))
			fmt.Printf("\n  Profile: %s\n", profileName)
			fmt.Printf("  Added:   %d new config(s)\n", inserted)
			fmt.Printf("  Skipped: %d duplicate(s)\n", len(configs)-inserted)
			fmt.Printf("  Total:   %d in profile\n\n", total)

			return nil
		}),
	}

	cmd.Flags().BoolVar(&testFirst, "test-first", false, "test configs before saving")
	cmd.Flags().StringVar(&profile, "profile", "", "profile name (default: last config name)")
	return cmd
}