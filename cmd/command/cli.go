package command

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/cmd/fs"
	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/internal/tester"
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

var (
	dbPath  string
	verbose bool
	root    *cobra.Command
)

type CLI struct {
	DBPath *string
}

func init() {
	root = &cobra.Command{
		Use:   "atabeh",
		Short: "VPN/proxy config manager with intelligent testing",
		Long: `Atabeh - Automated VPN/proxy configuration manager.

Features:
  - Universal config ingestion (subscriptions, files, raw URIs)
  - Intelligent fake ping detection
  - Stability scoring over time
  - VPN/Proxy connection with auto-selection
  - Export to multiple client formats
  - Profile-based organization

Examples:
  atabeh add https://subscription.url
  atabeh test --all --tag-reasons
  atabeh connect --auto --real-time
  atabeh status --watch
  atabeh export --profile 1 --format xray`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	root.PersistentFlags().StringVar(&dbPath, "db", fs.DBPath(), "database path")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if verbose {
			logger.SetLevel(logger.LevelDebug)
		}
		return nil
	}

	cli := &CLI{DBPath: &dbPath}

	root.AddCommand(
		cli.AddCommand(),
		cli.TestCommand(),
		cli.ListCommand(),
		cli.RemoveCommand(),
		cli.ExportCommand(),
		cli.RankCommand(),
		cli.ConnectCommand(),
		cli.DisconnectCommand(),
		cli.StatusCommand(),
		VersionCommand(),
	)
}

func Execute() error {
	return root.Execute()
}

func (c *CLI) WrapRepo(fn func(*repository.Repo, *cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		db, err := storage.Open(*c.DBPath)
		if err != nil {
			return err
		}
		defer db.Close()

		repo := repository.NewFromDB(db)
		return fn(repo, cmd, args)
	}
}

func testAndFilter(configs []*common.NormalizedConfig, cfg TestConfig) []*common.NormalizedConfig {
	testerCfg := tester.Config{
		Attempts:        cfg.Attempts,
		Timeout:         tester.DefaultTimeout,
		ConcurrentTests: cfg.Concurrent,
		StabilityWindow: tester.DefaultStabilityWindow,
	}

	if cfg.StabilityWindow > 0 {
		testerCfg.StabilityWindow = time.Duration(cfg.StabilityWindow)
	}

	results := tester.TestAll(configs, testerCfg)

	working := make([]*common.NormalizedConfig, 0)
	for i, r := range results {
		if r.Reachable {
			working = append(working, configs[i])
		}
	}

	return working
}
