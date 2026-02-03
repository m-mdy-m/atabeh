package logger

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"

	"github.com/m-mdy-m/atabeh/internal/common"
)

func ConfigReport(tag string, cfg *common.NormalizedConfig) {
	if !isEnabled(LevelInfo) {
		return
	}

	green := color.New(color.FgGreen, color.Bold)
	greenDim := color.New(color.FgGreen)

	green.Fprintf(os.Stdout, "[%s] [%-12s] ┌─ Config ────────────────────────\n", timestamp(), tag)

	fields := map[string]string{
		"name":      cfg.Name,
		"protocol":  string(cfg.Protocol),
		"server":    fmt.Sprintf("%s:%d", cfg.Server, cfg.Port),
		"transport": string(cfg.Transport),
		"security":  cfg.Security,
	}
	if cfg.UUID != "" {
		fields["uuid"] = truncate(cfg.UUID, 8) + "…"
	}
	if cfg.Password != "" {
		fields["password"] = strings.Repeat("•", min(len(cfg.Password), 8))
	}
	if cfg.Method != "" {
		fields["method"] = cfg.Method
	}

	keys := sortedKeys(fields)
	maxLen := longestKey(keys)

	for _, k := range keys {
		greenDim.Fprintf(os.Stdout, "  │  %-*s : %s\n", maxLen, k, fields[k])
	}

	if len(cfg.Extra) > 0 {
		greenDim.Fprintf(os.Stdout, "  │  %-*s : %s\n", maxLen, "extra", formatKV(cfg.Extra))
	}

	green.Fprintf(os.Stdout, "  └─────────────────────────────────\n")
}

func PingReport(tag string, name string, attempts, successes int, avgMs, minMs, maxMs int64) {
	if !isEnabled(LevelInfo) {
		return
	}

	loss := 0
	if attempts > 0 {
		loss = ((attempts - successes) * 100) / attempts
	}

	var statusStr string
	var c *color.Color
	switch {
	case successes == attempts && successes > 0:
		statusStr = "✓ REACHABLE"
		c = color.New(color.FgGreen, color.Bold)
	case successes == 0:
		statusStr = "✗ UNREACHABLE"
		c = color.New(color.FgRed, color.Bold)
	default:
		statusStr = "⚠ PARTIAL"
		c = color.New(color.FgYellow, color.Bold)
	}

	c.Fprintf(os.Stdout, "[%s] [%-12s] ┌─ Ping ──────────────────────────\n", timestamp(), tag)
	c.Fprintf(os.Stdout, "  │  target    : %s\n", name)
	c.Fprintf(os.Stdout, "  │  status    : %s\n", statusStr)
	c.Fprintf(os.Stdout, "  │  attempts  : %d/%d  (loss %d%%)\n", successes, attempts, loss)

	if successes > 0 {
		c.Fprintf(os.Stdout, "  │  latency   : avg %d ms / min %d ms / max %d ms\n", avgMs, minMs, maxMs)
	}
	c.Fprintf(os.Stdout, "  └─────────────────────────────────\n")
}

func SummaryReport(results []*common.PingResult) {
	if !isEnabled(LevelInfo) || len(results) == 0 {
		return
	}

	line := strings.Repeat("═", 62)
	white := color.New(color.FgWhite, color.Bold)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	dim := color.New(color.FgWhite)

	white.Fprintf(os.Stdout, "\n  %s\n", line)
	white.Fprintf(os.Stdout, "   atabeh — Test Summary (%d config(s))\n", len(results))
	white.Fprintf(os.Stdout, "  %s\n", line)

	reachable := 0
	var bestMs, worstMs int64
	bestMs = -1

	for i, r := range results {
		name := truncate(r.Config.Name, 22)

		if r.Reachable {
			reachable++
			green.Fprintf(os.Stdout, "   #%-2d  ✓  %-24s %4d ms   %-6s  %s:%d\n",
				i+1, name, r.AvgMs, r.Config.Protocol, r.Config.Server, r.Config.Port)
			if bestMs == -1 || r.AvgMs < bestMs {
				bestMs = r.AvgMs
			}
			if r.AvgMs > worstMs {
				worstMs = r.AvgMs
			}
		} else {
			red.Fprintf(os.Stdout, "   #%-2d  ✗  %-24s  —       %-6s  %s:%d\n",
				i+1, name, r.Config.Protocol, r.Config.Server, r.Config.Port)
		}
	}

	white.Fprintf(os.Stdout, "  %s\n", line)

	if bestMs == -1 {
		bestMs = 0
	}
	dim.Fprintf(os.Stdout, "   Reachable: %d/%d", reachable, len(results))
	if reachable > 0 {
		dim.Fprintf(os.Stdout, "  |  Best: %d ms  |  Worst: %d ms", bestMs, worstMs)
	}
	dim.Fprintf(os.Stdout, "\n")
	white.Fprintf(os.Stdout, "  %s\n\n", line)
}

func StorageReport(tag, source string, fetched, inserted, totalInDB int) {
	if !isEnabled(LevelInfo) {
		return
	}
	blue := color.New(color.FgCyan, color.Bold)
	dim := color.New(color.FgCyan)

	skipped := fetched - inserted
	blue.Fprintf(os.Stdout, "[%s] [%-12s] ┌─ Sync ──────────────────────────\n", timestamp(), tag)
	dim.Fprintf(os.Stdout, "  │  source     : %s\n", truncate(source, 52))
	dim.Fprintf(os.Stdout, "  │  fetched    : %d\n", fetched)
	dim.Fprintf(os.Stdout, "  │  inserted   : %d  (%d duplicate(s) skipped)\n", inserted, skipped)
	dim.Fprintf(os.Stdout, "  │  total in db: %d\n", totalInDB)
	blue.Fprintf(os.Stdout, "  └─────────────────────────────────\n")
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func longestKey(keys []string) int {
	n := 0
	for _, k := range keys {
		if len(k) > n {
			n = len(k)
		}
	}
	return n
}

func formatKV(m map[string]string) string {
	keys := sortedKeys(m)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, m[k]))
	}
	return strings.Join(parts, ", ")
}
