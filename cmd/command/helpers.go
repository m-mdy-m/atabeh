package command

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/tester"
	"github.com/m-mdy-m/atabeh/storage"
)

func printConfigTable(configs []*storage.ConfigRow) {
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

	fmt.Printf("\n  %d config(s)\n\n", len(configs))
}

func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func printTestResult(cfg *storage.ConfigRow, result *tester.Result, showTags bool) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if result.Reachable {
		fmt.Printf("  %s  %s  %d ms", green("✓"), cfg.Name, result.AvgMs)
		if showTags && result.StabilityScore > 0 {
			fmt.Printf("  (stability: %.1f)", result.StabilityScore)
		}
		fmt.Println()
	} else {
		fmt.Printf("  %s  %s  unreachable", red("✗"), cfg.Name)
		if showTags && result.FailureTag != "" {
			fmt.Printf("  (%s)", result.FailureTag)
		}
		fmt.Println()
	}
}

func printTestSummary(configs []*storage.ConfigRow, results []*tester.Result, showTags, showStability bool) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	alive := 0
	for _, r := range results {
		if r.Reachable {
			alive++
		}
	}

	fmt.Printf("\n  Results: %s/%d reachable\n", green(fmt.Sprintf("%d", alive)), len(results))
	fmt.Println("  " + strings.Repeat("-", 80))

	for i, r := range results {
		name := trunc(configs[i].Name, 28)
		if r.Reachable {
			info := fmt.Sprintf("%4d ms", r.AvgMs)
			if showStability && r.StabilityScore > 0 {
				info = fmt.Sprintf("%4d ms (s:%.1f)", r.AvgMs, r.StabilityScore)
			}
			fmt.Printf("  %s  %-28s  %s\n", green("✓"), name, info)
		} else {
			tag := ""
			if showTags && r.FailureTag != "" {
				tag = fmt.Sprintf(" [%s]", r.FailureTag)
			}
			fmt.Printf("  %s  %-28s  %s%s\n", red("✗"), name, "unreachable", tag)
		}
	}

	fmt.Println()
}
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func truncSource(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func toNormalized(s *storage.ConfigRow) *common.NormalizedConfig {
	return &common.NormalizedConfig{
		Name:      s.Name,
		Protocol:  s.Protocol,
		Server:    s.Server,
		Port:      s.Port,
		UUID:      s.UUID,
		Password:  s.Password,
		Method:    s.Method,
		Transport: s.Transport,
		Security:  s.Security,
	}
}
