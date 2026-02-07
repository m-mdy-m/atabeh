package command

import (
	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/connection"
	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) DisconnectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnect from VPN/proxy",
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			mgr := connection.NewManager(repo)

			logger.Infof("disconnect", "Disconnecting...")

			if err := mgr.Disconnect(); err != nil {
				logger.Errorf("disconnect", "Error: %v", err)
				return err
			}

			logger.Infof("disconnect", "✓ Disconnected")
			return nil
		}),
	}

	return cmd
}
