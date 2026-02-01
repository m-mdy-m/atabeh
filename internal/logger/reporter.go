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

	green.Fprintf(os.Stdout, "[%s] [%-12s] ┌── Config: %s\n", timestamp(), tag, cfg.Name)

	fields := map[string]string{
		"Protocol":  string(cfg.Protocol),
		"Server":    fmt.Sprintf("%s:%d", cfg.Server, cfg.Port),
		"Transport": string(cfg.Transport),
		"Security":  cfg.Security,
	}
	if cfg.UUID != "" {
		fields["UUID"] = cfg.UUID[:min(8, len(cfg.UUID))] + "…"
	}
	if cfg.Password != "" {
		fields["Password"] = strings.Repeat("•", min(len(cfg.Password), 8))
	}
	if cfg.Method != "" {
		fields["Method"] = cfg.Method
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	maxLen := 0
	for _, k := range keys {
		if len(k) > maxLen {
			maxLen = len(k)
		}
	}

	for _, k := range keys {
		greenDim.Fprintf(os.Stdout, "  │  %-*s : %s\n", maxLen, k, fields[k])
	}

	if len(cfg.Extra) > 0 {
		extraKeys := make([]string, 0, len(cfg.Extra))
		for k := range cfg.Extra {
			extraKeys = append(extraKeys, k)
		}
		sort.Strings(extraKeys)
		greenDim.Fprintf(os.Stdout, "  │  %-*s : ", maxLen, "Extra")
		parts := make([]string, 0, len(extraKeys))
		for _, k := range extraKeys {
			parts = append(parts, fmt.Sprintf("%s=%s", k, cfg.Extra[k]))
		}
		greenDim.Fprintf(os.Stdout, "%s\n", strings.Join(parts, ", "))
	}

	green.Fprintf(os.Stdout, "  └─────────────────────────\n")
}

func PingReport(tag string, configName string, attempts, successes int, avgMs, minMs, maxMs int64) {
	if !isEnabled(LevelInfo) {
		return
	}

	statusColor := color.New(color.FgGreen, color.Bold)
	switch {
	case successes == 0:
		statusColor = color.New(color.FgRed, color.Bold)
	case successes < attempts:
		statusColor = color.New(color.FgYellow, color.Bold)
	}

	lossPercent := 0
	if attempts > 0 {
		lossPercent = ((attempts - successes) * 100) / attempts
	}

	statusColor.Fprintf(os.Stdout, "[%s] [%-12s] ┌── Ping: %s\n", timestamp(), tag, configName)
	statusColor.Fprintf(os.Stdout, "  │  Attempts : %d/%d  (loss: %d%%)\n", successes, attempts, lossPercent)

	if successes > 0 {
		statusColor.Fprintf(os.Stdout, "  │  Avg      : %d ms\n", avgMs)
		statusColor.Fprintf(os.Stdout, "  │  Min      : %d ms\n", minMs)
		statusColor.Fprintf(os.Stdout, "  │  Max      : %d ms\n", maxMs)
	} else {
		statusColor.Fprintf(os.Stdout, "  │  Result   : UNREACHABLE\n")
	}
	statusColor.Fprintf(os.Stdout, "  └─────────────────────────\n")
}

func SummaryReport(results []*common.PingResult) {
	if !isEnabled(LevelInfo) || len(results) == 0 {
		return
	}

	line := strings.Repeat("═", 60)
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
		name := r.Config.Name
		if len(name) > 20 {
			name = name[:17] + "…"
		}

		if r.Reachable {
			reachable++
			green.Fprintf(os.Stdout, "   #%-2d  ✓  %-22s %4d ms   %-6s  %s:%d\n",
				i+1, name, r.AvgMs, r.Config.Protocol,
				r.Config.Server, r.Config.Port)
			if bestMs == -1 || r.AvgMs < bestMs {
				bestMs = r.AvgMs
			}
			if r.AvgMs > worstMs {
				worstMs = r.AvgMs
			}
		} else {
			red.Fprintf(os.Stdout, "   #%-2d  ✗  %-22s   —       %-6s  %s:%d\n",
				i+1, name, r.Config.Protocol,
				r.Config.Server, r.Config.Port)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
