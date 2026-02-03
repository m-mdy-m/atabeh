package main

import (
	"log"

	"github.com/m-mdy-m/atabeh/cmd"
	"github.com/m-mdy-m/atabeh/cmd/fs"
)

func main() {
	if err := fs.EnsureDirs(); err != nil {
		log.Fatalf("cannot setup atabeh dirs: %v", err)
	}

	// ۲. اجرا CLI
	cmd.Execute()
}
