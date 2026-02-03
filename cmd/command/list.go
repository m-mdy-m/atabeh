package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/storage"
)

func (c *CLI) ListCommand() *cobra.Command {
	var (
		listProtocol string
		aliveOnly    bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all stored configs",
		Long: `Displays every config in the database as a table.
Use --protocol to filter by a specific protocol (vless, vmess, ss, trojan, socks).
Use --alive to show only configs that passed their last connectivity test.

Examples:
  atabeh list
  atabeh list --protocol vless
  atabeh list --alive`,
		RunE: c.WrapRepo(func(repo *storage.ConfigRepo, cmd *cobra.Command, args []string) error {
			var (
				configs []*common.StoredConfig
				err     error
			)

			if aliveOnly {
				configs, err = repo.ListAlive()
			} else {
				configs, err = repo.List(common.Kind(listProtocol))
			}
			if err != nil {
				return err
			}

			if len(configs) == 0 {
				fmt.Println("No configs stored. Use `atabeh add` or `atabeh sync`.")
				return nil
			}

			printTable(configs)
			return nil
		}),
	}

	cmd.Flags().StringVar(
		&listProtocol,
		"protocol",
		"",
		"filter by protocol (vless|vmess|ss|trojan|socks)",
	)
	cmd.Flags().BoolVar(
		&aliveOnly,
		"alive",
		false,
		"show only configs that are alive",
	)

	return cmd
}
