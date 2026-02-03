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
	Short: "Atabeh — VPN/proxy config fetcher, tester & ranker",
	Long: `Atabeh collects VPN/proxy configs from subscriptions or manual input,
tests every one for real connectivity, ranks them by latency, and stores
the results in a local SQLite database.

Examples:
  atabeh add "vless://uuid@vpn.example.com:443?security=tls#Server1"
  atabeh sync https://sub.example.com/configs
  atabeh sub add https://raw.githubusercontent.com/user/repo/main/sub
  atabeh sub sync-all --test
  atabeh test --all
  atabeh rank
  atabeh list --alive`,
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
		cli.ListCommand(),
		cli.TestCommand(),
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
