package normalizer

import (
	"fmt"

	"github.com/m-mdy-m/atabeh/internal/common"
)

func Deduplicate(configs []*common.NormalizedConfig) []*common.NormalizedConfig {
	seen := make(map[string]struct{}, len(configs))
	unique := make([]*common.NormalizedConfig, 0, len(configs))

	for _, cfg := range configs {
		key := dedupKey(cfg)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, cfg)
	}

	return unique
}

func dedupKey(cfg *common.NormalizedConfig) string {
	switch cfg.Protocol {
	case common.Vless, common.VMess:
		return fmt.Sprintf("%s|%s|%d|%s|%s",
			cfg.Protocol, cfg.Server, cfg.Port, cfg.UUID, cfg.Transport)

	case common.Shadowsocks:
		return fmt.Sprintf("%s|%s|%d|%s|%s",
			cfg.Protocol, cfg.Server, cfg.Port, cfg.Password, cfg.Method)

	case common.Trojan:
		return fmt.Sprintf("%s|%s|%d|%s|%s",
			cfg.Protocol, cfg.Server, cfg.Port, cfg.Password, cfg.Transport)

	default:
		return fmt.Sprintf("%s|%s|%d",
			cfg.Protocol, cfg.Server, cfg.Port)
	}
}
