package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "0.2.0"
	GitCommit = "dev"
	BuildDate = "unknown"
)

func VersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("atabeh %s\n", Version)
			if GitCommit != "dev" {
				fmt.Printf("  commit: %s\n", GitCommit)
			}
			if BuildDate != "unknown" {
				fmt.Printf("  built:  %s\n", BuildDate)
			}
		},
	}
}
