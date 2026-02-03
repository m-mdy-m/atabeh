package command

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/storage"
	"github.com/m-mdy-m/atabeh/internal/tester"
)

func (c *CLI) TestCommand() *cobra.Command {
	var (
		testAll     bool
		testID      int
		testTimeout int
	)

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test connectivity of stored configs",
		Long: `Runs TCP-level connectivity tests against one or all stored configs
and persists the results.

Examples:
  atabeh test --all
  atabeh test --id 3
  atabeh test --all --timeout 10`,
		RunE: c.WrapRepo(func(repo *storage.ConfigRepo, cmd *cobra.Command, args []string) error {
			cfg := tester.Config{
				Attempts:        3,
				Timeout:         time.Duration(testTimeout) * time.Second,
				ConcurrentTests: 10,
				TestDelay:       100 * time.Millisecond,
			}

			switch {
			case testID > 0:
				return runSingle(repo, cfg, testID)
			case testAll:
				return runAll(repo, cfg)
			default:
				return fmt.Errorf("use --all or --id <N>")
			}
		}),
	}

	cmd.Flags().BoolVar(&testAll, "all", false, "test all stored configs")
	cmd.Flags().IntVar(&testID, "id", 0, "test a single config by id")
	cmd.Flags().IntVar(&testTimeout, "timeout", 5, "per-connection timeout in seconds")

	return cmd
}
