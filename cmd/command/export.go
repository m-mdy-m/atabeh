package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/exporter"
	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) ExportCommand() *cobra.Command {
	var (
		profileID int
		format    string
		output    string
		bestN     int
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export configs to various formats",
		Long: `Export configurations to client formats.

Supported formats:
  - sing-box: Sing-box configuration JSON
  - clash: Clash YAML (coming soon)
  - v2ray: V2Ray JSON (coming soon)

Examples:
  atabeh export --profile 1 --format sing-box
  atabeh export --profile 2 --format sing-box --best 5
  atabeh export --profile 1 --format sing-box --output config.json`,
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			if profileID == 0 {
				return fmt.Errorf("--profile required")
			}

			// Get profile
			profile, err := repo.GetProfile(profileID)
			if err != nil {
				return fmt.Errorf("profile: %w", err)
			}

			// Get configs
			configs, err := repo.ListConfigsByProfile(profileID)
			if err != nil {
				return err
			}

			if len(configs) == 0 {
				return fmt.Errorf("no configs in profile %d", profileID)
			}

			// Filter to best N if requested
			if bestN > 0 {
				if bestN < len(configs) {
					configs = configs[:bestN]
				}
			}

			logger.Infof("export", "exporting %d configs as %s", len(configs), format)

			// Export
			var data []byte
			switch format {
			case "sing-box":
				data, err = exporter.ToSingBox(configs)
			default:
				return fmt.Errorf("unsupported format: %s", format)
			}

			if err != nil {
				return fmt.Errorf("export: %w", err)
			}

			// Determine output path
			outPath := output
			if outPath == "" {
				outPath = fmt.Sprintf("%s.%s.json", profile.Name, format)
			}

			// Write file
			if err := os.WriteFile(outPath, data, 0644); err != nil {
				return fmt.Errorf("write: %w", err)
			}

			abs, _ := filepath.Abs(outPath)
			fmt.Printf("\n  ✓ Exported %d configs to:\n", len(configs))
			fmt.Printf("    %s\n\n", abs)

			return nil
		}),
	}

	cmd.Flags().IntVar(&profileID, "profile", 0, "profile to export (required)")
	cmd.Flags().StringVar(&format, "format", "sing-box", "export format")
	cmd.Flags().IntVar(&bestN, "best", 0, "export only N best configs")
	cmd.Flags().StringVar(&output, "output", "", "output file path")

	return cmd
}
