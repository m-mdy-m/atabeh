package logger

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func Info(m string) {
	fmt.Println(m)
}

func Infof(format string, args ...any) {
	Info(fmt.Sprintf(format, args...))
}

func Warning(m string) {
	color.Yellow("⚠ %s", m)
}

func Warningf(format string, args ...any) {
	Warning(fmt.Sprintf(format, args...))
}

func Error(m string) {
	color.New(color.FgRed).Fprintln(os.Stderr, "✗", m)
}

func Errorf(format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	Error(err.Error())
	return err
}
func Verbose(m string) {
	color.Cyan("→ %s", m)
}

func Verbosef(format string, args ...any) {
	Verbose(fmt.Sprintf(format, args...))
}

func Fatal(m string) {
	Error(m)
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	Fatal(fmt.Sprintf(format, args...))
}
