package command

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/storage"
)

func printTable(configs []*storage.ConfigRow) {
	if len(configs) == 0 {
		fmt.Println("  (empty)")
		return
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Printf("  %-4s  %-24s  %-8s  %-28s  %-5s  %-10s  %s\n",
		"ID", "Name", "Proto", "Server", "Port", "Status", "Latency")
	fmt.Println("  " + strings.Repeat("-", 90))

	for _, c := range configs {
		status := red("dead")
		latency := "—"
		if c.IsAlive {
			status = green("alive")
			latency = fmt.Sprintf("%d ms", c.LastPing)
		}
		fmt.Printf("  %-4d  %-24s  %-8s  %-28s  %-5d  %-10s  %s\n",
			c.ID, trunc(c.Name, 24), string(c.Protocol),
			trunc(c.Server, 28), c.Port, status, latency)
	}
	fmt.Printf("  %d config(s)\n", len(configs))
}

func printPingResult(cfg *storage.ConfigRow, result *common.PingResult) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if result.Reachable {
		fmt.Printf("  %s  %s  avg %d ms  min %d ms  max %d ms  (%d/%d ok)\n",
			green("✓"), cfg.Name, result.AvgMs, result.MinMs, result.MaxMs,
			result.Successes, result.Attempts)
	} else {
		fmt.Printf("  %s  %s  unreachable  (%d/%d ok)\n",
			red("✗"), cfg.Name, result.Successes, result.Attempts)
	}
}

func printTestSummary(storeds []*storage.ConfigRow, results []*common.PingResult) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	alive := 0
	for _, r := range results {
		if r.Reachable {
			alive++
		}
	}

	fmt.Printf("\n  Test results (%d/%d reachable)\n", alive, len(results))
	fmt.Println("  " + strings.Repeat("-", 70))

	for i, r := range results {
		name := trunc(storeds[i].Name, 24)
		if r.Reachable {
			fmt.Printf("  %s  %-24s  %4d ms  %s:%d\n",
				green("✓"), name, r.AvgMs, r.Config.Server, r.Config.Port)
		} else {
			fmt.Printf("  %s  %-24s  —       %s:%d\n",
				red("✗"), name, r.Config.Server, r.Config.Port)
		}
	}
	fmt.Println()
}

func printSyncSummary(source string, fetched, inserted, total int) {
	fmt.Printf("  source   %s\n", trunc(source, 60))
	fmt.Printf("  fetched  %d\n", fetched)
	fmt.Printf("  new      %d  (%d duplicate(s) skipped)\n", inserted, fetched-inserted)
	fmt.Printf("  total    %d in db\n", total)
}

func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
