package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/internal/tester"
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

type CLI struct {
	DBPath *string
}

func NewCLI(dbPath *string) *CLI {
	return &CLI{DBPath: dbPath}
}

func (c *CLI) WrapRepo(handler func(*repository.Repo, *cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		db, err := storage.Open(*c.DBPath)
		if err != nil {
			return err
		}
		defer db.Close()

		repo := repository.NewFromDB(db)

		return handler(repo, cmd, args)
	}
}

func runSingle(repo *repository.Repo, cfg tester.Config, id int) error {
	stored, err := repo.GetConfigByID(id)
	if err != nil {
		return fmt.Errorf("config id=%d: %w", id, err)
	}

	norm := toNormalized(stored)
	result := tester.Test(norm, cfg)

	if err := repo.UpdateConfigPingResult(id, result); err != nil {
		return err
	}
	printPingResult(stored, result)
	return nil
}

func runAll(repo *repository.Repo, cfg tester.Config) error {
	storeds, err := repo.ListConfigs("")
	if err != nil {
		return err
	}
	if len(storeds) == 0 {
		fmt.Println("  Nothing to test. Add configs first with `atabeh add` or `atabeh sync`.")
		return nil
	}

	norms := make([]*common.NormalizedConfig, len(storeds))
	for i, s := range storeds {
		norms[i] = toNormalized(s)
	}

	results := tester.TestAll(norms, cfg)

	for i, r := range results {
		if err := repo.UpdateConfigPingResult(storeds[i].ID, r); err != nil {
			logger.Errorf("test", "save result for id=%d: %v", storeds[i].ID, err)
		}
	}

	printTestSummary(storeds, results)
	return nil
}

func toNormalized(s *storage.ConfigRow) *common.NormalizedConfig {
	return &common.NormalizedConfig{
		Name: s.Name, Protocol: s.Protocol,
		Server: s.Server, Port: s.Port,
		UUID: s.UUID, Password: s.Password, Method: s.Method,
		Transport: s.Transport, Security: s.Security,
	}
}
