package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/cmd/command"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

var (
	DatabasePath string
	Verbose      bool
)

var Root = &cobra.Command{
	Use:   "atabeh",
	Short: "Atabeh — VPN/proxy config manager with concurrent testing",
	Long: `Atabeh collects, validates, tests, and ranks VPN/proxy configs.
Supports profiles for organization and concurrent testing for speed.

Examples:
  # Add configs with testing
  atabeh add "vless://..." --test-first
  
  # Sync subscription with concurrent testing  
  atabeh sync https://example.com/sub --test-first --concurrent 30
  
  # List profiles
  atabeh profile list
  
  # Test all configs concurrently
  atabeh test --all --concurrent 20
  
  # List configs grouped by profile
  atabeh list --grouped`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	Root.PersistentFlags().StringVar(&DatabasePath, "db", defaultDBPath(), "path to SQLite database")
	Root.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "enable debug logging")

	Root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if Verbose {
			logger.SetLevel(logger.LevelDebug)
		}
		return nil
	}

	cli := command.NewCLI(&DatabasePath)
	Root.AddCommand(
		cli.AddCommand(),
		cli.SyncCommand(),
		cli.TestCommand(),
		cli.ListCommand(),
		cli.ProfileCommand(),
		cli.RemoveCommand(),
		cli.RankCommand(),
		cli.SubCommand(),
		command.VersionCommand,
	)
}

func defaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return home + "/.config/atabeh/atabeh.db"
}

func Execute() {
	if err := Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
