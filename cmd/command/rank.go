package command

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) RankCommand() *cobra.Command {
	var (
		top        int
		withSource bool
	)

	cmd := &cobra.Command{
		Use:   "rank",
		Short: "Show configs ranked by latency",
		Long: `Displays all configs sorted by last test result:
alive configs first (lowest latency at the top), dead ones at the bottom.

Run ` + "`atabeh test --all`" + ` first to populate test results.

Examples:
  atabeh rank
  atabeh test --all && atabeh rank`,
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			configs, err := repo.ListConfigs("")
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

			if top > 0 && top < len(configs) {
				configs = configs[:top]
			}

			var profileSourceByID map[int]string
			if withSource {
				profiles, err := repo.ListProfiles()
				if err != nil {
					return err
				}
				profileSourceByID = make(map[int]string, len(profiles))
				for _, profile := range profiles {
					profileSourceByID[profile.ID] = profile.Source
				}
			}

			green := color.New(color.FgGreen).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()

			if withSource {
				fmt.Printf("  %-4s  %-24s  %-8s  %-28s  %s  %s\n",
					"#", "Name", "Proto", "Server", "Source", "Latency")
				fmt.Println("  " + strings.Repeat("-", 120))
			} else {
				fmt.Printf("  %-4s  %-24s  %-8s  %-28s  %s\n", "#", "Name", "Proto", "Server", "Latency")
				fmt.Println("  " + strings.Repeat("-", 75))
			}

			for i, c := range configs {
				name := trunc(c.Name, 24)
				server := trunc(c.Server, 28)
				source := ""
				if withSource {
					source = rawConfigURI(c.Extra)
					if source == "" {
						source = profileSourceByID[c.ProfileID]
					}
					if source == "" {
						source = c.Source
					}
				}
				if c.IsAlive {
					if withSource {
						fmt.Printf("  %-4d  %-24s  %-8s  %-28s  %s  %s\n",
							i+1, name, string(c.Protocol), server, source,
							green(fmt.Sprintf("%d ms", c.LastPing)))
					} else {
						fmt.Printf("  %-4d  %-24s  %-8s  %-28s  %s\n",
							i+1, name, string(c.Protocol), server,
							green(fmt.Sprintf("%d ms", c.LastPing)))
					}
				} else {
					if withSource {
						fmt.Printf("  %-4d  %-24s  %-8s  %-28s  %s  %s\n",
							i+1, name, string(c.Protocol), server, source,
							red("dead"))
					} else {
						fmt.Printf("  %-4d  %-24s  %-8s  %-28s  %s\n",
							i+1, name, string(c.Protocol), server,
							red("dead"))
					}
				}
			}
			fmt.Println()
			return nil
		}),
	}

	cmd.Flags().IntVarP(&top, "top", "n", 0, "show only the top N configs")
	cmd.Flags().BoolVar(&withSource, "source", false, "show profile source for each config")

	return cmd
}
func rawConfigURI(extraJSON string) string {
	if extraJSON == "" {
		return ""
	}
	var extra map[string]string
	if err := json.Unmarshal([]byte(extraJSON), &extra); err != nil {
		return ""
	}
	return extra["raw_uri"]
}
