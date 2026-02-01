package parsers

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

func init() { Register(&socksParser{}) }

type socksParser struct{}

func (s *socksParser) Protocol() common.Kind { return common.Socks }

// ParseURI parses: socks[4|5]://[user:pass@]host:port#name
func (s *socksParser) ParseURI(uri string) (*common.RawConfig, error) {
	logger.Debugf("socks", "parsing: %.100s", uri)

	// Normalise scheme to "socks" so url.Parse works uniformly
	scheme := extractScheme(uri)
	normalised := "socks" + strings.TrimPrefix(uri, scheme)

	u, err := url.Parse(normalised)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}

	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("missing server host")
	}

	port := 1080 // SOCKS default
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
	}

	extra := map[string]string{
		"scheme": strings.TrimSuffix(scheme, "://"),
	}

	username, password := "", ""
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()

		if password == "" && username != "" {
			if decoded, err := tryBase64(username); err == nil {
				parts := strings.SplitN(string(decoded), ":", 2)
				if len(parts) == 2 {
					username = parts[0]
					password = parts[1]
				}
			}
		}
	}

	if username != "" {
		extra["username"] = username
	}

	cfg := &common.RawConfig{
		Protocol:  common.Socks,
		Name:      u.Fragment,
		Server:    host,
		Port:      port,
		Password:  password,
		Transport: common.TCP,
		Security:  "none",
		Extra:     extra,
	}

	logger.Debugf("socks", "→ name=%q server=%s:%d auth=%v",
		cfg.Name, cfg.Server, cfg.Port, username != "")
	return cfg, nil
}

// extractScheme returns "socks://", "socks4://", or "socks5://" from the URI.
func extractScheme(uri string) string {
	for _, prefix := range []string{"socks5://", "socks4://", "socks://"} {
		if strings.HasPrefix(uri, prefix) {
			return prefix
		}
	}
	return "socks://"
}
