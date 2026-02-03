package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/internal/normalizer"
	"github.com/m-mdy-m/atabeh/internal/parsers"
	"github.com/m-mdy-m/atabeh/internal/storage"
)

func (c *CLI) AddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <uri>",
		Short: "Add a single VPN/proxy config URI",
		Long: `Parses the given URI (vless://, vmess://, ss://, trojan://, socks://),
validates it, and stores it in the local database.

Example:
  atabeh add "vless://uuid@vpn.example.com:443?security=tls#MyServer"`,
		Args: cobra.ExactArgs(1),
		RunE: c.WrapRepo(func(repo *storage.ConfigRepo, cmd *cobra.Command, args []string) error {
			uri := args[0]

			raw, err := parsers.ParseAll([]string{uri})
			if err != nil {
				return fmt.Errorf("parse failed: %w", err)
			}
			if len(raw) == 0 {
				return fmt.Errorf("could not parse the provided URI")
			}

			configs, err := normalizer.Normalize(raw)
			if err != nil {
				return fmt.Errorf("normalisation failed: %w", err)
			}
			if len(configs) == 0 {
				return fmt.Errorf("config failed validation")
			}

			id, inserted, err := repo.InsertOrSkip(configs[0], "manual")
			if err != nil {
				return err
			}

			if !inserted {
				logger.Infof("add", "duplicate — already stored as id=%d", id)
				fmt.Printf("Already stored (id=%d)\n", id)
				return nil
			}

			logger.ConfigReport("add", configs[0])
			fmt.Printf("Added config id=%d  name=%q\n", id, configs[0].Name)
			return nil
		}),
	}
}
