package command

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) ListCommand() *cobra.Command {
	var (
		profiles  bool
		protocol  string
		aliveOnly bool
		profileID int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configs or profiles",
		Long: `List configurations or profiles with various filters.

Examples:
  atabeh list                    # list all configs
  atabeh list --profiles         # list profiles
  atabeh list --profile 1        # list configs in profile 1
  atabeh list --alive            # list only working configs
  atabeh list --protocol vless   # filter by protocol`,
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			if profiles {
				return listProfiles(repo)
			}

			if profileID > 0 {
				return listByProfile(repo, profileID)
			}

			return listConfigs(repo, protocol, aliveOnly)
		}),
	}

	cmd.Flags().BoolVar(&profiles, "profiles", false, "list profiles instead of configs")
	cmd.Flags().IntVar(&profileID, "profile", 0, "list configs in specific profile")
	cmd.Flags().StringVar(&protocol, "protocol", "", "filter by protocol")
	cmd.Flags().BoolVar(&aliveOnly, "alive", false, "show only alive configs")

	return cmd
}

func listProfiles(repo *repository.Repo) error {
	profiles, err := repo.ListProfiles()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		fmt.Println("  No profiles. Use `atabeh add` to create one.")
		return nil
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Printf("\n  %-4s  %-30s  %-12s  %-8s  %-8s\n",
		"ID", "Name", "Type", "Configs", "Alive")
	fmt.Println("  " + strings.Repeat("-", 70))

	for _, p := range profiles {
		name := trunc(p.Name, 30)
		typeColor := cyan
		if p.Type == "subscription" {
			typeColor = yellow
		}

		aliveStr := fmt.Sprintf("%d", p.AliveCount)
		if p.AliveCount > 0 {
			aliveStr = green(aliveStr)
		}

		fmt.Printf("  %-4d  %-30s  %-12s  %-8d  %-8s\n",
			p.ID, name, typeColor(p.Type), p.ConfigCount, aliveStr)
	}

	fmt.Printf("\n  %d profile(s)\n\n", len(profiles))
	return nil
}

func listByProfile(repo *repository.Repo, profileID int) error {
	profile, err := repo.GetProfile(profileID)
	if err != nil {
		return err
	}

	fmt.Printf("\n  Profile: %s\n", profile.Name)
	fmt.Printf("  Type:    %s\n", profile.Type)
	fmt.Printf("  Source:  %s\n\n", trunc(profile.Source, 60))

	configs, err := repo.ListConfigsByProfile(profileID)
	if err != nil {
		return err
	}

	if len(configs) == 0 {
		fmt.Println("  No configs in this profile")
		return nil
	}

	printConfigTable(configs)
	return nil
}

func listConfigs(repo *repository.Repo, protocol string, aliveOnly bool) error {
	var configs []*storage.ConfigRow
	var err error

	if aliveOnly {
		configs, err = repo.ListAliveConfigs()
	} else {
		configs, err = repo.ListConfigs(common.Kind(protocol))
	}

	if err != nil {
		return err
	}

	if len(configs) == 0 {
		fmt.Println("  No configs. Use `atabeh add` first.")
		return nil
	}

	printConfigTable(configs)
	return nil
}
