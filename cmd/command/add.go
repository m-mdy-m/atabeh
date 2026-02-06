package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/internal/normalizer"
	"github.com/m-mdy-m/atabeh/internal/parsers"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) AddCommand() *cobra.Command {
	var (
		testFirst  bool
		testConfig TestConfig
		profile    string
	)

	cmd := &cobra.Command{
		Use:   "add <source>",
		Short: "Add configs from any source (URL, file, or raw text)",
		Long: `Universal config ingestion - handles subscriptions, files, and raw configs.

Source types:
  - Subscription URL (http://... or https://...)
  - Local file (@path/to/file.txt)
  - Raw config URIs or mixed content

Examples:
  atabeh add https://example.com/subscription
  atabeh add @configs.txt
  atabeh add "vless://..." --test-first
  atabeh add https://... --test-first --attempts 5`,
		Args: cobra.MinimumNArgs(1),
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing source argument")
			}
			source := args[0]

			if strings.HasPrefix(source, "@") {
				content, err := os.ReadFile(source[1:])
				if err != nil {
					return fmt.Errorf("read file: %w", err)
				}
				source = string(content)
			}

			logger.Infof("add", "processing: %s", truncSource(source, 60))

			rawConfigs, err := parsers.FetchAndParseAll(source)
			if err != nil {
				return fmt.Errorf("fetch/parse: %w", err)
			}

			if len(rawConfigs) == 0 {
				return fmt.Errorf("no configs found")
			}

			logger.Infof("add", "fetched %d raw configs", len(rawConfigs))

			configs, err := normalizer.Normalize(rawConfigs)
			if err != nil {
				return fmt.Errorf("normalize: %w", err)
			}

			logger.Infof("add", "normalized to %d unique configs", len(configs))

			if len(configs) == 0 {
				return fmt.Errorf("all configs invalid or duplicates")
			}

			if testFirst {
				logger.Infof("add", "testing before save...")
				configs = testAndFilter(configs, testConfig)
				if len(configs) == 0 {
					return fmt.Errorf("no configs passed tests")
				}
				logger.Infof("add", "%d configs passed", len(configs))
			}

			profileName := profile
			if profileName == "" {
				profileName = normalizer.ExtractProfileName(source)
			}

			profileSource := source
			if isURL(source) {
				profileSource = source
			} else {
				profileSource = "manual"
			}

			profileType := "manual"
			if isURL(source) {
				profileType = "subscription"
			}

			profileID, err := repo.GetOrCreateProfile(profileName, profileSource, profileType)
			if err != nil {
				return fmt.Errorf("profile: %w", err)
			}

			inserted, err := repo.InsertConfigBatch(configs, profileID)
			if err != nil {
				return fmt.Errorf("insert: %w", err)
			}

			total, _ := repo.CountConfigsByProfile(int(profileID))

			fmt.Printf("\n  Profile:  %s\n", profileName)
			fmt.Printf("  Type:     %s\n", profileType)
			fmt.Printf("  Added:    %d new\n", inserted)
			fmt.Printf("  Skipped:  %d duplicates\n", len(configs)-inserted)
			fmt.Printf("  Total:    %d configs\n\n", total)

			return nil
		}),
	}

	cmd.Flags().BoolVar(&testFirst, "test-first", false, "test before saving")
	cmd.Flags().StringVar(&profile, "profile", "", "custom profile name")
	cmd.Flags().IntVar(&testConfig.Attempts, "attempts", 3, "test attempts per config")
	cmd.Flags().IntVar(&testConfig.Concurrent, "concurrent", 20, "concurrent tests")
	cmd.Flags().IntVar(&testConfig.TimeoutSec, "timeout", 5, "timeout in seconds")
	cmd.Flags().IntVar(&testConfig.StabilityWindow, "stability-window", 0, "stability test duration (sec)")

	return cmd
}
