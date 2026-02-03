package command

import (
	"fmt"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/internal/storage"
	"github.com/m-mdy-m/atabeh/internal/tester"
	"github.com/spf13/cobra"
)

type CLI struct {
	DBPath *string
}

func NewCLI(dbPath *string) *CLI { return &CLI{DBPath: dbPath} }

func (c *CLI) WrapRepo(handler func(*storage.ConfigRepo, *cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		db, err := storage.Open(*c.DBPath)
		if err != nil {
			return err
		}
		defer db.Close()
		repo := storage.NewConfigRepo(db)
		return handler(repo, cmd, args)
	}
}
func runSingle(repo *storage.ConfigRepo, cfg tester.Config, id int) error {
	stored, err := repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("config id=%d: %w", id, err)
	}

	norm := storedToNormalized(stored)
	result := tester.Test(norm, cfg)

	return repo.UpdatePingResult(id, result)
}

func runAll(repo *storage.ConfigRepo, cfg tester.Config) error {
	storeds, err := repo.List("")
	if err != nil {
		return err
	}
	if len(storeds) == 0 {
		fmt.Println("Nothing to test. Add configs first with `atabeh add` or `atabeh sync`.")
		return nil
	}

	norms := make([]*common.NormalizedConfig, len(storeds))
	for i, s := range storeds {
		norms[i] = storedToNormalized(s)
	}

	results := tester.TestAll(norms, cfg)

	for i, r := range results {
		if err := repo.UpdatePingResult(storeds[i].ID, r); err != nil {
			logger.Errorf("test", "failed to save result for id=%d: %v", storeds[i].ID, err)
		}
	}

	logger.SummaryReport(results)
	return nil
}

func storedToNormalized(s *common.StoredConfig) *common.NormalizedConfig {
	return &common.NormalizedConfig{
		Name:      s.Name,
		Protocol:  s.Protocol,
		Server:    s.Server,
		Port:      s.Port,
		UUID:      s.UUID,
		Password:  s.Password,
		Method:    s.Method,
		Transport: s.Transport,
		Security:  s.Security,
	}
}
