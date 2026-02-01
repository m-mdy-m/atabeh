package normalizer

import (
	"fmt"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

const tag = "normalizer"

func Normalize(raw []*common.RawConfig) ([]*common.NormalizedConfig, error) {
	logger.Infof(tag, "normalizing %d raw config(s)", len(raw))

	seen := map[string]bool{} // dedup key -> already seen
	var result []*common.NormalizedConfig

	for i, r := range raw {
		cfg, err := normalizeOne(r)
		if err != nil {
			logger.Warnf(tag, "[%d] skipping invalid config: %v", i, err)
			continue
		}

		key := dedupKey(cfg)
		if seen[key] {
			logger.Debugf(tag, "[%d] duplicate config, skipping: %s", i, key)
			continue
		}
		seen[key] = true

		result = append(result, cfg)
		logger.Debugf(tag, "[%d] normalized OK: %s", i, cfg.Name)
	}

	logger.Infof(tag, "normalized %d -> %d unique config(s)", len(raw), len(result))
	return result, nil
}

func normalizeOne(r *common.RawConfig) (*common.NormalizedConfig, error) {
	if r.Server == "" {
		return nil, fmt.Errorf("missing server")
	}
	if r.Port <= 0 || r.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", r.Port)
	}

	switch r.Protocol {
	case common.Vless, common.VMess:
		if r.UUID == "" {
			return nil, fmt.Errorf("missing UUID for %s", r.Protocol)
		}
	case common.Shadowsocks:
		if r.Password == "" {
			return nil, fmt.Errorf("missing password for shadowsocks")
		}
		if r.Method == "" {
			return nil, fmt.Errorf("missing encryption method for shadowsocks")
		}
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", r.Protocol)
	}

	name := cleanName(r.Name)
	if name == "" {
		name = fmt.Sprintf("%s-%s", r.Protocol, r.Server)
	}

	transport := r.Transport
	if transport == "" {
		transport = common.TCP
	}

	security := r.Security
	if security == "" {
		security = "none"
	}

	cfg := &common.NormalizedConfig{
		Name:      name,
		Protocol:  r.Protocol,
		Server:    r.Server,
		Port:      r.Port,
		UUID:      r.UUID,
		Password:  r.Password,
		Method:    r.Method,
		Transport: transport,
		Security:  security,
		Extra:     r.Extra,
	}

	logger.Debugf(tag, "normalized: name=%q proto=%s server=%s:%d",
		cfg.Name, cfg.Protocol, cfg.Server, cfg.Port)

	return cfg, nil
}

func cleanName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.Trim(name, "«»‹›「」【】〔〕（）()[]{}⟨⟩")
	name = strings.TrimSpace(name)
	return name
}

func dedupKey(c *common.NormalizedConfig) string {
	switch c.Protocol {
	case common.Vless, common.VMess:
		return fmt.Sprintf("%s|%s|%d|%s", c.Protocol, c.Server, c.Port, c.UUID)
	case common.Shadowsocks:
		return fmt.Sprintf("%s|%s|%d|%s", c.Protocol, c.Server, c.Port, c.Password)
	default:
		return fmt.Sprintf("%s|%s|%d", c.Protocol, c.Server, c.Port)
	}
}
