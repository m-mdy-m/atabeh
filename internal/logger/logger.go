package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var (
	currentLevel Level = LevelInfo
	mu           sync.Mutex
	timeFormat   = "15:04:05"
)

func SetLevel(l Level) {
	mu.Lock()
	defer mu.Unlock()
	currentLevel = l
}

func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	default:
		return LevelInfo
	}
}

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func isEnabled(l Level) bool {
	mu.Lock()
	defer mu.Unlock()
	return l >= currentLevel
}

func timestamp() string {
	return time.Now().Format(timeFormat)
}

func Debug(tag, msg string) {
	if !isEnabled(LevelDebug) {
		return
	}
	color.New(color.FgCyan).Fprintf(os.Stderr, "[%s] [DEBUG] [%s] %s\n", timestamp(), tag, msg)
}
func Debugf(tag, format string, args ...any) {
	Debug(tag, fmt.Sprintf(format, args...))
}
func Info(tag, msg string) {
	if !isEnabled(LevelInfo) {
		return
	}
	color.New(color.FgWhite).Fprintf(os.Stdout, "[%s] [INFO ] [%s] %s\n", timestamp(), tag, msg)
}
func Infof(tag, format string, args ...any) {
	Info(tag, fmt.Sprintf(format, args...))
}

func Warn(tag, msg string) {
	if !isEnabled(LevelWarn) {
		return
	}
	color.New(color.FgYellow).Fprintf(os.Stderr, "[%s] [WARN ] [%s] %s\n", timestamp(), tag, msg)
}

func Warnf(tag, format string, args ...any) {
	Warn(tag, fmt.Sprintf(format, args...))
}

func Error(tag, msg string) {
	if !isEnabled(LevelError) {
		return
	}
	color.New(color.FgRed).Fprintf(os.Stderr, "[%s] [ERROR] [%s] %s\n", timestamp(), tag, msg)
}

func Errorf(tag, format string, args ...any) {
	Error(tag, fmt.Sprintf(format, args...))
}

func Fatal(tag, msg string) {
	color.New(color.FgRed, color.Bold).Fprintf(os.Stderr, "[%s] [FATAL] [%s] %s\n", timestamp(), tag, msg)
	os.Exit(1)
}

func Fatalf(tag, format string, args ...any) {
	Fatal(tag, fmt.Sprintf(format, args...))
}

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
