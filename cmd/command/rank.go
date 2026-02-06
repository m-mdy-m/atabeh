package command

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) RankCommand() *cobra.Command {
	var (
		topN        int
		byStability bool
		showSource  bool
	)

	cmd := &cobra.Command{
		Use:   "rank",
		Short: "Show configs ranked by performance",
		Long: `Display configs ranked by latency or stability.

Run 'atabeh test --all' first to populate test results.

Examples:
  atabeh rank                    # rank by latency
  atabeh rank --top 10           # show top 10
  atabeh rank --by-stability     # rank by stability score
  atabeh rank --source           # show config source`,
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			configs, err := repo.ListConfigs("")
			if err != nil {
				return err
			}

			if len(configs) == 0 {
				fmt.Println("  No configs. Use `atabeh add` first.")
				return nil
			}

			// Sort: alive first (by latency), then dead
			sort.SliceStable(configs, func(i, j int) bool {
				a, b := configs[i], configs[j]

				if a.IsAlive != b.IsAlive {
					return a.IsAlive
				}

				if a.IsAlive {
					return a.LastPing < b.LastPing
				}

				return false
			})

			// Limit to top N
			if topN > 0 && topN < len(configs) {
				configs = configs[:topN]
			}

			green := color.New(color.FgGreen).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()

			fmt.Printf("\n  %-4s  %-28s  %-8s  %-28s  %s\n",
				"#", "Name", "Proto", "Server", "Latency")
			fmt.Println("  " + strings.Repeat("-", 85))

			for i, cfg := range configs {
				name := trunc(cfg.Name, 28)
				server := trunc(cfg.Server, 28)

				if cfg.IsAlive {
					fmt.Printf("  %-4d  %-28s  %-8s  %-28s  %s\n",
						i+1, name, string(cfg.Protocol), server,
						green(fmt.Sprintf("%d ms", cfg.LastPing)))
				} else {
					fmt.Printf("  %-4d  %-28s  %-8s  %-28s  %s\n",
						i+1, name, string(cfg.Protocol), server,
						red("dead"))
				}
			}

			fmt.Println()
			return nil
		}),
	}

	cmd.Flags().IntVarP(&topN, "top", "n", 0, "show top N configs")
	cmd.Flags().BoolVar(&byStability, "by-stability", false, "rank by stability score")
	cmd.Flags().BoolVar(&showSource, "source", false, "show config source")

	return cmd
}
