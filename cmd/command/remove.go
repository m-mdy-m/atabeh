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
		removeProfile int
		removeAll     bool
		confirm       bool
	)

	cmd := &cobra.Command{
		Use:   "remove [id]",
		Short: "Remove config(s) or profile(s) from database",
		Long: `Remove one or more configs, entire profiles, or everything.

Examples:
  atabeh remove 3                    # Remove config ID 3
  atabeh remove --profile 2          # Remove entire profile 2
  atabeh remove --all                # Remove everything
  atabeh remove --all --yes          # Remove without confirmation`,
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			switch {
			case removeAll:
				return removeAllConfigs(repo, confirm)

			case removeProfile > 0:
				return removeProfileAndConfigs(repo, removeProfile, confirm)

			case len(args) == 1:
				id, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid config id %q: must be integer", args[0])
				}
				return removeSingleConfig(repo, id, confirm)

			default:
				return fmt.Errorf("provide a config id, --profile <ID>, or --all")
			}
		}),
	}

	cmd.Flags().IntVar(&removeProfile, "profile", 0, "remove entire profile and all its configs")
	cmd.Flags().BoolVar(&removeAll, "all", false, "remove all configs and profiles")
	cmd.Flags().BoolVarP(&confirm, "yes", "y", false, "skip confirmation prompt")
	return cmd
}

func removeSingleConfig(repo *repository.Repo, id int, skipConfirm bool) error {
	cfg, err := repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("config id=%d not found: %w", id, err)
	}
	profile, err := repo.GetProfile(cfg.ProfileID)
	if err != nil {
		return fmt.Errorf("get profile: %w", err)
	}

	if !skipConfirm {
		fmt.Printf("  Remove config:\n")
		fmt.Printf("    ID:       %d\n", cfg.ID)
		fmt.Printf("    Name:     %s\n", cfg.Name)
		fmt.Printf("    Server:   %s:%d\n", cfg.Server, cfg.Port)
		fmt.Printf("    Profile:  %s\n", profile.Name)
		fmt.Printf("\n  Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("  Cancelled.")
			return nil
		}
	}

	if err := repo.DeleteByID(id); err != nil {
		return err
	}

	fmt.Printf("  ✓ Removed config id=%d (%s)\n", id, cfg.Name)

	// Show updated profile stats
	updatedProfile, err := repo.GetProfile(cfg.ProfileID)
	if err == nil {
		fmt.Printf("  Profile '%s' now has %d config(s)\n",
			updatedProfile.Name, updatedProfile.ConfigCount)
	}

	return nil
}

func removeProfileAndConfigs(repo *repository.Repo, profileID int, skipConfirm bool) error {
	// Get profile info first
	profile, err := repo.GetProfile(profileID)
	if err != nil {
		return fmt.Errorf("profile id=%d not found: %w", profileID, err)
	}

	if !skipConfirm {
		fmt.Printf("  Remove profile:\n")
		fmt.Printf("    ID:       %d\n", profile.ID)
		fmt.Printf("    Name:     %s\n", profile.Name)
		fmt.Printf("    Type:     %s\n", profile.Type)
		fmt.Printf("    Configs:  %d total, %d alive\n",
			profile.ConfigCount, profile.AliveCount)
		fmt.Printf("\n  This will delete the profile AND all %d config(s)!\n",
			profile.ConfigCount)
		fmt.Printf("  Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("  Cancelled.")
			return nil
		}
	}

	// Delete profile (CASCADE will delete all configs)
	if err := repo.DeleteProfile(profileID); err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}

	fmt.Printf("  ✓ Removed profile '%s' and %d config(s)\n",
		profile.Name, profile.ConfigCount)

	// Show remaining stats
	profiles, err := repo.ListProfiles()
	if err == nil {
		fmt.Printf("  %d profile(s) remaining\n", len(profiles))
	}

	return nil
}

func removeAllConfigs(repo *repository.Repo, skipConfirm bool) error {
	// Get current stats
	profiles, err := repo.ListProfiles()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		fmt.Println("  Database is already empty.")
		return nil
	}

	totalConfigs, _ := repo.Count()

	if !skipConfirm {
		fmt.Printf("  ⚠️  WARNING: This will delete EVERYTHING!\n\n")
		fmt.Printf("  Profiles: %d\n", len(profiles))
		fmt.Printf("  Configs:  %d\n", totalConfigs)
		fmt.Printf("\n  This action cannot be undone!\n")
		fmt.Printf("  Type 'DELETE ALL' to confirm: ")

		var response string
		fmt.Scanln(&response)
		if response != "DELETE ALL" {
			fmt.Println("  Cancelled.")
			return nil
		}
	}

	// Delete all profiles (CASCADE will delete all configs)
	for _, profile := range profiles {
		if err := repo.DeleteProfile(profile.ID); err != nil {
			return fmt.Errorf("delete profile %d: %w", profile.ID, err)
		}
	}

	fmt.Printf("  ✓ Deleted %d profile(s) and %d config(s)\n",
		len(profiles), totalConfigs)
	fmt.Println("  Database is now empty.")

	return nil
}
