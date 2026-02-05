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
		Short: "Remove config(s) or profile(s)",
		Long: `Remove configs, profiles, or everything.

Examples:
  atabeh remove 3                  # Remove config ID 3
  atabeh remove --profile 2        # Remove profile 2 (+ all its configs via CASCADE)
  atabeh remove --all              # Remove everything
  atabeh remove --all --yes        # Skip confirmation`,
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			switch {
			case removeAll:
				return handleRemoveAll(repo, confirm)
			case removeProfile > 0:
				return handleRemoveProfile(repo, removeProfile, confirm)
			case len(args) == 1:
				id, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid config id: %q", args[0])
				}
				return handleRemoveConfig(repo, id, confirm)
			default:
				return fmt.Errorf("provide config id, --profile <ID>, or --all")
			}
		}),
	}

	cmd.Flags().IntVar(&removeProfile, "profile", 0, "remove entire profile + configs")
	cmd.Flags().BoolVar(&removeAll, "all", false, "remove everything")
	cmd.Flags().BoolVarP(&confirm, "yes", "y", false, "skip confirmation")
	return cmd
}

func handleRemoveConfig(repo *repository.Repo, id int, skipConfirm bool) error {
	cfg, err := repo.GetConfigByID(id)
	if err != nil {
		return fmt.Errorf("config %d not found: %w", id, err)
	}

	profile, err := repo.GetProfile(cfg.ProfileID)
	if err != nil {
		return fmt.Errorf("get profile: %w", err)
	}

	if !skipConfirm {
		fmt.Printf("  Remove config:\n")
		fmt.Printf("    ID:      %d\n", cfg.ID)
		fmt.Printf("    Name:    %s\n", cfg.Name)
		fmt.Printf("    Server:  %s:%d\n", cfg.Server, cfg.Port)
		fmt.Printf("    Profile: %s\n", profile.Name)
		fmt.Printf("\n  Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("  Cancelled.")
			return nil
		}
	}

	if err := repo.DeleteConfigByID(id); err != nil {
		return err
	}

	fmt.Printf("  ✓ Removed config %d (%s)\n", id, cfg.Name)

	// Show updated stats
	updatedProfile, err := repo.GetProfile(cfg.ProfileID)
	if err == nil {
		fmt.Printf("  Profile '%s' now has %d config(s)\n",
			updatedProfile.Name, updatedProfile.ConfigCount)
	}

	return nil
}

func handleRemoveProfile(repo *repository.Repo, profileID int, skipConfirm bool) error {
	profile, err := repo.GetProfile(profileID)
	if err != nil {
		return fmt.Errorf("profile %d not found: %w", profileID, err)
	}

	if !skipConfirm {
		fmt.Printf("  Remove profile:\n")
		fmt.Printf("    ID:      %d\n", profile.ID)
		fmt.Printf("    Name:    %s\n", profile.Name)
		fmt.Printf("    Type:    %s\n", profile.Type)
		fmt.Printf("    Configs: %d total, %d alive\n",
			profile.ConfigCount, profile.AliveCount)
		fmt.Printf("\n  ⚠️  CASCADE DELETE: all %d config(s) will be removed!\n",
			profile.ConfigCount)
		fmt.Printf("  Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("  Cancelled.")
			return nil
		}
	}

	if err := repo.DeleteProfile(profileID); err != nil {
		return err
	}

	fmt.Printf("  ✓ Removed profile '%s' and %d config(s)\n",
		profile.Name, profile.ConfigCount)

	// Show remaining
	profiles, err := repo.ListProfiles()
	if err == nil {
		fmt.Printf("  %d profile(s) remaining\n", len(profiles))
	}

	return nil
}

func handleRemoveAll(repo *repository.Repo, skipConfirm bool) error {
	profiles, err := repo.ListProfiles()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		fmt.Println("  Database is already empty.")
		return nil
	}

	totalConfigs, _ := repo.CountConfigs()

	if !skipConfirm {
		fmt.Printf("  ⚠️  WARNING: This will delete EVERYTHING!\n\n")
		fmt.Printf("  Profiles: %d\n", len(profiles))
		fmt.Printf("  Configs:  %d\n", totalConfigs)
		fmt.Printf("\n  This action cannot be undone!\n")
		fmt.Printf("  Are you sure? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" {
			fmt.Println("  Cancelled.")
			return nil
		}
	}

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
