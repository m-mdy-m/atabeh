package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var VersionCommand = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("atabeh %s (commit %s, built %s)\n", Version, GitCommit, BuildDate)
	},
}
