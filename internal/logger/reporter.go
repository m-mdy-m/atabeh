package logger

import (
	"os"
	"strings"

	"github.com/fatih/color"
)

func ConfigReport(tag string, fields map[string]string) {
	if !isEnabled(LevelInfo) {
		return
	}
	color.New(color.FgGreen, color.Bold).Fprintf(os.Stdout, "[%s] [%s] ─── Config Report ───\n", timestamp(), tag)
	maxLen := 0
	for k := range fields {
		if len(k) > maxLen {
			maxLen = len(k)
		}
	}
	for k, v := range fields {
		padded := k + strings.Repeat(" ", maxLen-len(k))
		color.New(color.FgGreen).Fprintf(os.Stdout, "  [%s]   %-s : %s\n", timestamp(), padded, v)
	}
	color.New(color.FgGreen, color.Bold).Fprintf(os.Stdout, "[%s] [%s] ─────────────────────\n", timestamp(), tag)
}

func PingReport(tag string, configName string, attempts int, successes int, avgMs int64, minMs int64, maxMs int64) {
	if !isEnabled(LevelInfo) {
		return
	}
	status := color.New(color.FgGreen, color.Bold)
	if successes == 0 {
		status = color.New(color.FgRed, color.Bold)
	} else if successes < attempts {
		status = color.New(color.FgYellow, color.Bold)
	}

	lossPercent := 0
	if attempts > 0 {
		lossPercent = ((attempts - successes) * 100) / attempts
	}

	status.Fprintf(os.Stdout, "[%s] [%s] ┌── Ping: %s\n", timestamp(), tag, configName)
	status.Fprintf(os.Stdout, "  │  Attempts : %d/%d  (loss: %d%%)\n", successes, attempts, lossPercent)
	if successes > 0 {
		status.Fprintf(os.Stdout, "  │  Avg      : %d ms\n", avgMs)
		status.Fprintf(os.Stdout, "  │  Min      : %d ms\n", minMs)
		status.Fprintf(os.Stdout, "  │  Max      : %d ms\n", maxMs)
	} else {
		status.Fprintf(os.Stdout, "  │  Result   : UNREACHABLE\n")
	}
	status.Fprintf(os.Stdout, "  └─────────────────────────\n")
}
