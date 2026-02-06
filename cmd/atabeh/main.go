package main

import (
	"log"
	"os"

	"github.com/m-mdy-m/atabeh/cmd/command"
	"github.com/m-mdy-m/atabeh/cmd/fs"
)

func main() {
	if err := fs.EnsureDirs(); err != nil {
		log.Fatalf("setup failed: %v", err)
	}

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
