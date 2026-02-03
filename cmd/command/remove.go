package command

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/storage"
)

func (c *CLI) RemoveCommand() *cobra.Command {
	var (
		removeSource string // --source
		removeAll    bool   // --all
	)

	cmd := &cobra.Command{
		Use:   "remove [id]",
		Short: "Remove config(s) from the database",
		Long: `Removes one or more configs.

  atabeh remove 3
  atabeh remove --source https://…
  atabeh remove --all`,
		RunE: c.WrapRepo(func(repo *storage.ConfigRepo, cmd *cobra.Command, args []string) error {
			switch {
			// --all
			case removeAll:
				if err := repo.Clear(); err != nil {
					return err
				}
				fmt.Println("All configs removed.")

			// --source <url>
			case removeSource != "":
				n, err := repo.DeleteBySource(removeSource)
				if err != nil {
					return err
				}
				fmt.Printf("Removed %d config(s) from source %q\n", n, removeSource)

			// positional id
			case len(args) == 1:
				id, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid id %q: must be integer", args[0])
				}
				if err := repo.DeleteByID(id); err != nil {
					return err
				}
				fmt.Printf("Removed config id=%d\n", id)

			default:
				return fmt.Errorf("provide an id, --source, or --all")
			}
			return nil
		}),
	}

	cmd.Flags().StringVar(&removeSource, "source", "", "remove all configs from this subscription URL")
	cmd.Flags().BoolVar(&removeAll, "all", false, "remove every config (destructive!)")

	return cmd
}
