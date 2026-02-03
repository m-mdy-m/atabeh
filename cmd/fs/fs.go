package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

func BaseDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "atabeh")
}

func EnsureBaseDir() error {
	dir := BaseDir()
	return os.MkdirAll(dir, 0o755)
}

func DBPath() string {
	return filepath.Join(BaseDir(), "atabeh.db")
}
func EnsureDBDir() error {
	dir := filepath.Dir(DBPath())
	return os.MkdirAll(dir, 0o755)
}
func EnsureFileDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0o755)
}
func EnsureDirs() error {
	dirs := []string{
		BaseDir(),
		filepath.Join(BaseDir(), "cache"),
		filepath.Join(BaseDir(), "logs"),
		filepath.Join(BaseDir(), "subscriptions"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("cannot create dir %q: %w", d, err)
		}
	}
	return nil
}
