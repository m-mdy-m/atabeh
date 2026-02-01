package parsers

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

func init() { Register(&trojanParser{}) }

type trojanParser struct{}

func (t *trojanParser) Protocol() common.Kind { return common.Trojan }

// ParseURI parses: trojan://password@host:port?type=…&security=…#name
func (t *trojanParser) ParseURI(uri string) (*common.RawConfig, error) {
	logger.Debugf("trojan", "parsing: %.100s", uri)

	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}
	if u.Scheme != "trojan" {
		return nil, fmt.Errorf("expected trojan scheme, got %s", u.Scheme)
	}

	password := ""
	if u.User != nil {
		password = u.User.Username()
	}
	if password == "" {
		return nil, fmt.Errorf("missing password")
	}
	password, err = url.PathUnescape(password)
	if err != nil {
		return nil, fmt.Errorf("password decode: %w", err)
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
		security = "tls"
	}

	extra := extractExtra(params, "type", "security")

	cfg := &common.RawConfig{
		Protocol:  common.Trojan,
		Name:      u.Fragment,
		Server:    host,
		Port:      port,
		Password:  password,
		Transport: transport,
		Security:  security,
		Extra:     extra,
	}

	logger.Debugf("trojan", "→ name=%q server=%s:%d security=%s transport=%s",
		cfg.Name, cfg.Server, cfg.Port, cfg.Security, cfg.Transport)
	return cfg, nil
}
