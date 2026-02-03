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
	mu           sync.Mutex
	currentLevel = LevelInfo
	timeFormat   = "15:04:05"
)

func SetLevel(l Level) { mu.Lock(); currentLevel = l; mu.Unlock() }
func GetLevel() Level  { mu.Lock(); defer mu.Unlock(); return currentLevel }

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

var levelColors = map[Level]*color.Color{
	LevelDebug: color.New(color.FgCyan),
	LevelInfo:  color.New(color.FgWhite),
	LevelWarn:  color.New(color.FgYellow),
	LevelError: color.New(color.FgRed),
	LevelFatal: color.New(color.FgRed, color.Bold),
}

func enabled(l Level) bool { mu.Lock(); defer mu.Unlock(); return l >= currentLevel }

func emit(lvl Level, tag, msg string) {
	w := os.Stdout
	if lvl >= LevelWarn {
		w = os.Stderr
	}
	c := levelColors[lvl]
	ts := time.Now().Format(timeFormat)
	c.Fprintf(w, "%s  %-5s  %-12s  %s\n", ts, lvl, tag, msg)
}

func Debug(tag, msg string)               { if enabled(LevelDebug) { emit(LevelDebug, tag, msg) } }
func Debugf(tag, format string, a ...any) { if enabled(LevelDebug) { emit(LevelDebug, tag, fmt.Sprintf(format, a...)) } }
func Info(tag, msg string)                { if enabled(LevelInfo) { emit(LevelInfo, tag, msg) } }
func Infof(tag, format string, a ...any)  { if enabled(LevelInfo) { emit(LevelInfo, tag, fmt.Sprintf(format, a...)) } }
func Warn(tag, msg string)                { if enabled(LevelWarn) { emit(LevelWarn, tag, msg) } }
func Warnf(tag, format string, a ...any)  { if enabled(LevelWarn) { emit(LevelWarn, tag, fmt.Sprintf(format, a...)) } }
func Error(tag, msg string)               { if enabled(LevelError) { emit(LevelError, tag, msg) } }
func Errorf(tag, format string, a ...any) { if enabled(LevelError) { emit(LevelError, tag, fmt.Sprintf(format, a...)) } }
func Fatal(tag, msg string)               { emit(LevelFatal, tag, msg); os.Exit(1) }
func Fatalf(tag, format string, a ...any) { Fatal(tag, fmt.Sprintf(format, a...)) }