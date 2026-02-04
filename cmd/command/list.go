package command

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) ListCommand() *cobra.Command {
	var (
		proto     string
		aliveOnly bool
		profileID int
		grouped   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configs or profiles",
		Long: `Shows configs, optionally grouped by profile.

Examples:
  atabeh list
  atabeh list --grouped
  atabeh list --profile 1
  atabeh list --protocol vless
  atabeh list --alive`,
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			if grouped {
				return listGrouped(repo)
			}

			if profileID > 0 {
				return listByProfile(repo, profileID)
			}

			var configs []*storage.ConfigRow
			var err error

			if aliveOnly {
				configs, err = repo.ListAliveConfigs()
			} else {
				configs, err = repo.ListConfigs(common.Kind(proto))
			}

			if err != nil {
				return err
			}

			if len(configs) == 0 {
				fmt.Println("  No configs. Use `atabeh add` or `atabeh sync`.")
				return nil
			}

			printTable(configs)
			return nil
		}),
	}

	cmd.Flags().StringVar(&proto, "protocol", "", "filter by protocol")
	cmd.Flags().BoolVar(&aliveOnly, "alive", false, "show only alive configs")
	cmd.Flags().IntVar(&profileID, "profile", 0, "show configs from specific profile")
	cmd.Flags().BoolVar(&grouped, "grouped", false, "group configs by profile")
	return cmd
}

func listGrouped(repo *repository.Repo) error {
	profiles, err := repo.ListProfiles()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		fmt.Println("  No profiles.")
		return nil
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	for _, profile := range profiles {
		fmt.Printf("\n  %s %s %s\n",
			cyan("━━━━━ Profile:"),
			yellow(profile.Name),
			cyan("━━━━━"))

		fmt.Printf("  Type: %s  |  Configs: %d  |  Alive: %s\n",
			profile.Type, profile.ConfigCount, green(fmt.Sprintf("%d", profile.AliveCount)))

		configs, err := repo.ListConfigsByProfile(profile.ID)
		if err != nil {
			fmt.Printf("  Error loading configs: %v\n", err)
			continue
		}

		if len(configs) == 0 {
			fmt.Println("  (no configs)")
			continue
		}

		printTableCompact(configs)
	}

	fmt.Println()
	return nil
}

func listByProfile(repo *repository.Repo, profileID int) error {
	profile, err := repo.GetProfile(profileID)
	if err != nil {
		return err
	}

	fmt.Printf("\n  Profile: %s\n", profile.Name)
	fmt.Printf("  Type:    %s\n", profile.Type)
	fmt.Println()

	configs, err := repo.ListConfigsByProfile(profileID)
	if err != nil {
		return err
	}

	if len(configs) == 0 {
		fmt.Println("  No configs in this profile.")
		return nil
	}

	printTable(configs)
	return nil
}

func printTableCompact(configs []*storage.ConfigRow) {
	if len(configs) == 0 {
		return
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	for _, c := range configs {
		status := red("✗")
		latency := "dead"
		if c.IsAlive {
			status = green("✓")
			latency = fmt.Sprintf("%d ms", c.LastPing)
		}

		fmt.Printf("  %s  %-20s  %-8s  %-25s  %s\n",
			status, trunc(c.Name, 20), string(c.Protocol),
			trunc(c.Server, 25), latency)
	}
}
