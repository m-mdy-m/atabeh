package command

import (
	"fmt"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
)

func printTable(configs []*common.StoredConfig) {
	header := fmt.Sprintf(" %-4s  %-24s  %-8s  %-30s  %-5s  %-6s  %s",
		"ID", "Name", "Proto", "Server", "Port", "Status", "Latency")
	sep := strings.Repeat("─", len(header)+2)

	fmt.Println(sep)
	fmt.Println(header)
	fmt.Println(sep)

	for _, c := range configs {
		status := "✗ dead"
		latency := " —"
		if c.IsAlive {
			status = "✓ alive"
			latency = fmt.Sprintf("%d ms", c.LastPing)
		}

		name := truncStr(c.Name, 24)
		server := truncStr(c.Server, 30)

		fmt.Printf(" %-4d  %-24s  %-8s  %-30s  %-5d  %-7s  %s\n",
			c.ID, name, string(c.Protocol), server, c.Port, status, latency)
	}

	fmt.Println(sep)
	fmt.Printf(" Total: %d config(s)\n", len(configs))
}

func truncStr(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
