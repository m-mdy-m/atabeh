package command

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/storage"
)

func (c *CLI) RankCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rank",
		Short: "Show configs ranked by latency",
		Long: `Displays all configs sorted by last test result:
alive configs first (lowest latency at the top), dead ones at the bottom.

Run ` + "`atabeh test --all`" + ` first to populate test results.

Examples:
  atabeh rank
  atabeh test --all && atabeh rank`,
		RunE: c.WrapRepo(func(repo *storage.ConfigRepo, cmd *cobra.Command, args []string) error {
			configs, err := repo.List("")
			if err != nil {
				return err
			}
			if len(configs) == 0 {
				fmt.Println("  No configs. Use `atabeh add` or `atabeh sync`.")
				return nil
			}

			// sort: alive first (ascending latency), then dead
			sort.SliceStable(configs, func(i, j int) bool {
				a, b := configs[i], configs[j]
				if a.IsAlive != b.IsAlive {
					return a.IsAlive // alive before dead
				}
				if a.IsAlive {
					return a.LastPing < b.LastPing
				}
				return false
			})

			green := color.New(color.FgGreen).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()

			fmt.Printf("  %-4s  %-24s  %-8s  %-28s  %s\n", "#", "Name", "Proto", "Server", "Latency")
			fmt.Println("  " + strings.Repeat("-", 75))

			for i, c := range configs {
				name := trunc(c.Name, 24)
				server := trunc(c.Server, 28)
				if c.IsAlive {
					fmt.Printf("  %-4d  %-24s  %-8s  %-28s  %s\n",
						i+1, name, string(c.Protocol), server,
						green(fmt.Sprintf("%d ms", c.LastPing)))
				} else {
					fmt.Printf("  %-4d  %-24s  %-8s  %-28s  %s\n",
						i+1, name, string(c.Protocol), server,
						red("dead"))
				}
			}
			fmt.Println()
			return nil
		}),
	}
}
