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

func GetLevel() Level {
	mu.Lock()
	defer mu.Unlock()
	return currentLevel
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
		return "?????"
	}
}

func isEnabled(l Level) bool {
	mu.Lock()
	defer mu.Unlock()
	return l >= currentLevel
}

func timestamp() string { return time.Now().Format(timeFormat) }

func emit(lvl Level, w *os.File, c *color.Color, tag, msg string) {
	c.Fprintf(w, "[%s] [%s] [%-12s] %s\n", timestamp(), lvl, tag, msg)
}

func Debug(tag, msg string) {
	if !isEnabled(LevelDebug) {
		return
	}
	emit(LevelDebug, os.Stderr, color.New(color.FgCyan), tag, msg)
}

func Debugf(tag, format string, a ...any) { Debug(tag, fmt.Sprintf(format, a...)) }

func Info(tag, msg string) {
	if !isEnabled(LevelInfo) {
		return
	}
	emit(LevelInfo, os.Stdout, color.New(color.FgWhite), tag, msg)
}

func Infof(tag, format string, a ...any) { Info(tag, fmt.Sprintf(format, a...)) }

func Warn(tag, msg string) {
	if !isEnabled(LevelWarn) {
		return
	}
	emit(LevelWarn, os.Stderr, color.New(color.FgYellow), tag, msg)
}

func Warnf(tag, format string, a ...any) { Warn(tag, fmt.Sprintf(format, a...)) }

func Error(tag, msg string) {
	if !isEnabled(LevelError) {
		return
	}
	emit(LevelError, os.Stderr, color.New(color.FgRed), tag, msg)
}

func Errorf(tag, format string, a ...any) { Error(tag, fmt.Sprintf(format, a...)) }

func Fatal(tag, msg string) {
	emit(LevelFatal, os.Stderr, color.New(color.FgRed, color.Bold), tag, msg)
	os.Exit(1)
}

func Fatalf(tag, format string, a ...any) { Fatal(tag, fmt.Sprintf(format, a...)) }
