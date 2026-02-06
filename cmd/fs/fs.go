package fs

import (
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

func DBPath() string {
	return filepath.Join(BaseDir(), "atabeh.db")
}

func EnsureDirs() error {
	dirs := []string{
		BaseDir(),
		filepath.Join(BaseDir(), "cache"),
		filepath.Join(BaseDir(), "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
