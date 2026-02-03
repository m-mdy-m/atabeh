package command

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/storage"
)

func (c *CLI) RemoveCommand() *cobra.Command {
	var (
		source    string
		removeAll bool
	)

	cmd := &cobra.Command{
		Use:   "remove [id]",
		Short: "Remove config(s) from the database",
		Long: `Remove one or more configs.

Examples:
  atabeh remove 3
  atabeh remove --source https://sub.example.com/configs
  atabeh remove --all`,
		RunE: c.WrapRepo(func(repo *storage.ConfigRepo, cmd *cobra.Command, args []string) error {
			switch {
			case removeAll:
				if err := repo.Clear(); err != nil {
					return err
				}
				fmt.Println("  all configs removed")

			case source != "":
				n, err := repo.DeleteBySource(source)
				if err != nil {
					return err
				}
				fmt.Printf("  removed %d config(s) from %q\n", n, source)

			case len(args) == 1:
				id, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid id %q: must be integer", args[0])
				}
				if err := repo.DeleteByID(id); err != nil {
					return err
				}
				fmt.Printf("  removed id=%d\n", id)

			default:
				return fmt.Errorf("provide an id, --source, or --all")
			}
			return nil
		}),
	}

	cmd.Flags().StringVar(&source, "source", "", "remove all configs from this subscription URL")
	cmd.Flags().BoolVar(&removeAll, "all", false, "remove every config")
	return cmd
}
