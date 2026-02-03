package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/storage"
)

func (c *CLI) ListCommand() *cobra.Command {
	var (
		proto     string
		aliveOnly bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all stored configs",
		Long: `Shows every config in the database as a table.

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
				configs, err = repo.List(common.Kind(proto))
			}
			if err != nil {
				return err
			}
			if len(configs) == 0 {
				fmt.Println("  No configs stored. Use `atabeh add` or `atabeh sync`.")
				return nil
			}
			printTable(configs)
			return nil
		}),
	}

	cmd.Flags().StringVar(&proto, "protocol", "", "filter by protocol (vless|vmess|ss|trojan)")
	cmd.Flags().BoolVar(&aliveOnly, "alive", false, "show only configs that are alive")
	return cmd
}
