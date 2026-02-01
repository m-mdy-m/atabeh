package parsers

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

func init() { Register(&vlessParser{}) }

type vlessParser struct{}

func (v *vlessParser) Protocol() common.Kind { return common.Vless }

// ParseURI parses: vless://UUID@host:port?type=…&security=…&…#name
func (v *vlessParser) ParseURI(uri string) (*common.RawConfig, error) {
	logger.Debugf("vless", "parsing: %.100s", uri)

	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}
	if u.Scheme != "vless" {
		return nil, fmt.Errorf("expected vless scheme, got %s", u.Scheme)
	}

	uuid := ""
	if u.User != nil {
		uuid = u.User.Username()
	}
	if uuid == "" {
		return nil, fmt.Errorf("missing UUID")
	}

	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("missing server host")
	}

	port := 443
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
	}

	params := u.Query()

	transport := common.Kind(params.Get("type"))
	if transport == "" {
		transport = common.TCP
	}

	security := params.Get("security")
	if security == "" {
		security = "none"
	}

	extra := extractExtra(params, "type", "security")

	cfg := &common.RawConfig{
		Protocol:  common.Vless,
		Name:      u.Fragment,
		Server:    host,
		Port:      port,
		UUID:      uuid,
		Transport: transport,
		Security:  security,
		Extra:     extra,
	}

	logger.Debugf("vless", "→ name=%q server=%s:%d security=%s transport=%s",
		cfg.Name, cfg.Server, cfg.Port, cfg.Security, cfg.Transport)
	return cfg, nil
}
func extractExtra(params url.Values, skip ...string) map[string]string {
	skipSet := make(map[string]bool, len(skip))
	for _, s := range skip {
		skipSet[s] = true
	}
	extra := map[string]string{}
	for key, vals := range params {
		if skipSet[key] || len(vals) == 0 {
			continue
		}
		extra[key] = vals[0]
	}
	return extra
}
