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

	seen := map[string]bool{}
	var result []*common.NormalizedConfig

	for i, r := range raw {
		cfg, err := normalizeOne(r)
		if err != nil {
			logger.Warnf(tag, "[%d] skipped: %v", i, err)
			continue
		}

		key := dedupKey(cfg)
		if seen[key] {
			logger.Debugf(tag, "[%d] duplicate, skipping: %s", i, key)
			continue
		}
		seen[key] = true

		result = append(result, cfg)
		logger.Debugf(tag, "[%d] OK: %s", i, cfg.Name)
	}

	logger.Infof(tag, "normalized %d → %d unique config(s)", len(raw), len(result))
	return result, nil
}

func normalizeOne(r *common.RawConfig) (*common.NormalizedConfig, error) {
	if r.Server == "" {
		return nil, fmt.Errorf("missing server")
	}
	if r.Port <= 0 || r.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", r.Port)
	}

	if err := validateCredentials(r); err != nil {
		return nil, err
	}
	name := cleanName(r.Name)
	if name == "" {
		name = fallbackName(r)
	}

	transport := r.Transport
	if transport == "" {
		transport = common.TCP
	}

	security := r.Security
	if security == "" {
		security = defaultSecurity(r.Protocol)
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

func validateCredentials(r *common.RawConfig) error {
	switch r.Protocol {
	case common.Vless, common.VMess:
		if r.UUID == "" {
			return fmt.Errorf("missing UUID for %s", r.Protocol)
		}
	case common.Shadowsocks:
		if r.Password == "" {
			return fmt.Errorf("missing password for shadowsocks")
		}
		if r.Method == "" {
			return fmt.Errorf("missing encryption method for shadowsocks")
		}
	case common.Trojan:
		if r.Password == "" {
			return fmt.Errorf("missing password for trojan")
		}
	case common.Socks:
		// SOCKS allows anonymous connections — no credentials required
	default:
		return fmt.Errorf("unsupported protocol: %s", r.Protocol)
	}
	return nil
}

func defaultSecurity(proto common.Kind) string {
	if proto == common.Trojan {
		return "tls"
	}
	return "none"
}

func fallbackName(r *common.RawConfig) string {
	if r.Server == "" {
		return "atabeh-unknown"
	}
	return fmt.Sprintf("%s-%s", r.Protocol, r.Server)
}

func cleanName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.Trim(name, "«»‹›「」【】〔〕（）()[]{}⟨⟩")
	name = strings.TrimSpace(name)
	return name
}

func dedupKey(c *common.NormalizedConfig) string {
	cred := c.UUID
	if cred == "" {
		cred = c.Password
	}
	return fmt.Sprintf("%s|%s|%d|%s", c.Protocol, c.Server, c.Port, cred)
}
