package command

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) ProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage configuration profiles",
		Long: `Profile management for organizing configs.

Examples:
  atabeh profile list
  atabeh profile show 1
  atabeh profile delete 2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		c.profileListCmd(),
		c.profileShowCmd(),
		c.profileDeleteCmd(),
	)
	return cmd
}

func (c *CLI) profileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all profiles",
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			profiles, err := repo.ListProfiles()
			if err != nil {
				return err
			}
			if len(profiles) == 0 {
				fmt.Println("  No profiles. Add configs or sync subscriptions to create profiles.")
				return nil
			}

			green := color.New(color.FgGreen).SprintFunc()
			yellow := color.New(color.FgYellow).SprintFunc()
			cyan := color.New(color.FgCyan).SprintFunc()

			fmt.Printf("  %-4s  %-30s  %-12s  %-8s  %-8s\n",
				"ID", "Name", "Type", "Configs", "Alive")
			fmt.Println("  " + strings.Repeat("-", 70))

			for _, p := range profiles {
				name := trunc(p.Name, 30)
				configStr := fmt.Sprintf("%d", p.ConfigCount)
				aliveStr := fmt.Sprintf("%d", p.AliveCount)

				if p.AliveCount > 0 {
					aliveStr = green(aliveStr)
				}

				typeColor := cyan
				if p.Type == "subscription" {
					typeColor = yellow
				}

				fmt.Printf("  %-4d  %-30s  %-12s  %-8s  %-8s\n",
					p.ID, name, typeColor(p.Type), configStr, aliveStr)
			}
			fmt.Printf("\n  %d profile(s)\n", len(profiles))
			return nil
		}),
	}
}

func (c *CLI) profileShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show configs in a profile",
		Args:  cobra.ExactArgs(1),
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			var id int
			if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
				return fmt.Errorf("invalid profile ID")
			}

			profile, err := repo.GetProfile(id)
			if err != nil {
				return err
			}

			fmt.Printf("\n  Profile: %s\n", profile.Name)
			fmt.Printf("  Type:    %s\n", profile.Type)
			fmt.Printf("  Source:  %s\n", trunc(profile.Source, 60))
			fmt.Println()

			configs, err := repo.ListConfigsByProfile(id)
			if err != nil {
				return err
			}

			if len(configs) == 0 {
				fmt.Println("  No configs in this profile.")
				return nil
			}

			printTable(configs)
			return nil
		}),
	}
}

func (c *CLI) profileDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a profile and all its configs",
		Args:  cobra.ExactArgs(1),
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			var id int
			if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
				return fmt.Errorf("invalid profile ID")
			}

			profile, err := repo.GetProfile(id)
			if err != nil {
				return err
			}

			fmt.Printf("  Delete profile '%s' and all its %d config(s)? [y/N]: ",
				profile.Name, profile.ConfigCount)

			var response string
			fmt.Scanln(&response)

			if strings.ToLower(response) != "y" {
				fmt.Println("  Cancelled.")
				return nil
			}

			if err := repo.DeleteProfile(id); err != nil {
				return err
			}

			fmt.Printf("  Profile '%s' deleted.\n", profile.Name)
			return nil
		}),
	}
}
