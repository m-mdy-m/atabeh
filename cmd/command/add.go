package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/normalizer"
	"github.com/m-mdy-m/atabeh/internal/parsers"
	"github.com/m-mdy-m/atabeh/internal/storage"
)

func (c *CLI) AddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <uri>",
		Short: "Add a single VPN/proxy config URI",
		Long: `Parses the given URI and stores it in the local database.

Supported schemes: vless://, vmess://, ss://, trojan://,

Examples:
  atabeh add "vless://uuid@vpn.example.com:443?security=tls#MyServer"
  atabeh add "trojan://password@server.com:443#Iran1"`,
		Args: cobra.ExactArgs(1),
		RunE: c.WrapRepo(func(repo *storage.ConfigRepo, cmd *cobra.Command, args []string) error {
			raw, err := parsers.ParseText(args[0])
			if err != nil {
				return fmt.Errorf("parse: %w", err)
			}
			if len(raw) == 0 {
				return fmt.Errorf("could not parse the provided URI")
			}

			configs, err := normalizer.Normalize(raw)
			if err != nil {
				return fmt.Errorf("normalize: %w", err)
			}
			if len(configs) == 0 {
				return fmt.Errorf("config failed validation")
			}

			id, inserted, err := repo.InsertOrSkip(configs[0], "manual")
			if err != nil {
				return err
			}
			if !inserted {
				fmt.Printf("  already stored (id=%d)\n", id)
				return nil
			}
			fmt.Printf("  added  id=%d  name=%q  %s:%d\n", id, configs[0].Name, configs[0].Server, configs[0].Port)
			return nil
		}),
	}
}
