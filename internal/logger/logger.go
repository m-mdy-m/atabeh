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
	color.New(color.FgCyan).Fprintf(os.Stderr, "[%s] [DEBUG] [%-12s] %s\n", timestamp(), tag, msg)
}
func Debugf(tag, format string, args ...any) { Debug(tag, fmt.Sprintf(format, args...)) }

func Info(tag, msg string) {
	if !isEnabled(LevelInfo) {
		return
	}
	color.New(color.FgWhite).Fprintf(os.Stdout, "[%s] [INFO ] [%-12s] %s\n", timestamp(), tag, msg)
}
func Infof(tag, format string, args ...any) { Info(tag, fmt.Sprintf(format, args...)) }

func Warn(tag, msg string) {
	if !isEnabled(LevelWarn) {
		return
	}
	color.New(color.FgYellow).Fprintf(os.Stderr, "[%s] [WARN ] [%-12s] %s\n", timestamp(), tag, msg)
}
func Warnf(tag, format string, args ...any) { Warn(tag, fmt.Sprintf(format, args...)) }

func Error(tag, msg string) {
	if !isEnabled(LevelError) {
		return
	}
	color.New(color.FgRed).Fprintf(os.Stderr, "[%s] [ERROR] [%-12s] %s\n", timestamp(), tag, msg)
}
func Errorf(tag, format string, args ...any) { Error(tag, fmt.Sprintf(format, args...)) }

func Fatal(tag, msg string) {
	color.New(color.FgRed, color.Bold).Fprintf(os.Stderr, "[%s] [FATAL] [%-12s] %s\n", timestamp(), tag, msg)
	os.Exit(1)
}
func Fatalf(tag, format string, args ...any) { Fatal(tag, fmt.Sprintf(format, args...)) }
