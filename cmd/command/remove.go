package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) RemoveCommand() *cobra.Command {
	var (
		profileID int
		all       bool
		yes       bool
	)

	cmd := &cobra.Command{
		Use:   "remove [id]",
		Short: "Remove config(s) or profile(s)",
		Long: `Remove configurations or entire profiles.

Examples:
  atabeh remove 5              # remove config ID 5
  atabeh remove --profile 2    # remove profile 2 (cascade)
  atabeh remove --all          # remove everything
  atabeh remove --all --yes    # skip confirmation`,
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			switch {
			case all:
				return removeAll(repo, yes)
			case profileID > 0:
				return removeProfile(repo, profileID, yes)
			case len(args) == 1:
				id, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid config ID: %q", args[0])
				}
				return removeConfig(repo, id, yes)
			default:
				return fmt.Errorf("specify config ID, --profile <ID>, or --all")
			}
		}),
	}

	cmd.Flags().IntVar(&profileID, "profile", 0, "remove profile (cascade delete)")
	cmd.Flags().BoolVar(&all, "all", false, "remove everything")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation")

	return cmd
}

func removeConfig(repo *repository.Repo, id int, skip bool) error {
	cfg, err := repo.GetConfigByID(id)
	if err != nil {
		return fmt.Errorf("config %d: %w", id, err)
	}

	if !skip {
		fmt.Printf("  Remove config %d (%s - %s:%d)? [y/N]: ",
			id, cfg.Name, cfg.Server, cfg.Port)

		var resp string
		fmt.Scanln(&resp)
		if strings.ToLower(resp) != "y" {
			fmt.Println("  Cancelled")
			return nil
		}
	}

	if err := repo.DeleteConfigByID(id); err != nil {
		return err
	}

	fmt.Printf("  ✓ Removed config %d\n", id)
	return nil
}

func removeProfile(repo *repository.Repo, id int, skip bool) error {
	profile, err := repo.GetProfile(id)
	if err != nil {
		return fmt.Errorf("profile %d: %w", id, err)
	}

	if !skip {
		fmt.Printf("  Remove profile '%s' and %d config(s)? [y/N]: ",
			profile.Name, profile.ConfigCount)

		var resp string
		fmt.Scanln(&resp)
		if strings.ToLower(resp) != "y" {
			fmt.Println("  Cancelled")
			return nil
		}
	}

	if err := repo.DeleteProfile(id); err != nil {
		return err
	}

	fmt.Printf("  ✓ Removed profile '%s' and %d config(s)\n",
		profile.Name, profile.ConfigCount)
	return nil
}

func removeAll(repo *repository.Repo, skip bool) error {
	profiles, err := repo.ListProfiles()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		fmt.Println("  Already empty")
		return nil
	}

	total, _ := repo.CountConfigs()

	if !skip {
		fmt.Printf("  ⚠️  Remove ALL data? (profiles: %d, configs: %d) [y/N]: ",
			len(profiles), total)

		var resp string
		fmt.Scanln(&resp)
		if strings.ToLower(resp) != "y" {
			fmt.Println("  Cancelled")
			return nil
		}
	}

	for _, p := range profiles {
		if err := repo.DeleteProfile(p.ID); err != nil {
			return err
		}
	}

	fmt.Printf("  ✓ Removed %d profile(s) and %d config(s)\n", len(profiles), total)
	return nil
}
