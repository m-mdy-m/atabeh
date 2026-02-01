package parsers

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

func init() { Register(&ssParser{}) }

type ssParser struct{}

func (s *ssParser) Protocol() common.Kind { return common.Shadowsocks }

// ParseURI handles both SIP002 and legacy Shadowsocks URI formats.
// SIP002:  ss://base64(method:password)@host:port#name
// Legacy:  ss://base64(method:password@host:port)#name
func (s *ssParser) ParseURI(uri string) (*common.RawConfig, error) {
	logger.Debugf("ss", "parsing: %.100s", uri)

	cfg, err := s.parseSIP002(uri)
	if err != nil {
		logger.Debugf("ss", "SIP002 failed (%v), trying legacy", err)
		cfg, err = s.parseLegacy(uri)
		if err != nil {
			return nil, fmt.Errorf("both SIP002 and legacy failed: %w", err)
		}
	}
	return cfg, nil
}

func (s *ssParser) parseSIP002(uri string) (*common.RawConfig, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}
	if u.Scheme != "ss" {
		return nil, fmt.Errorf("expected ss scheme")
	}
	if u.User == nil {
		return nil, fmt.Errorf("missing userinfo")
	}

	decoded, err := tryBase64(u.User.Username())
	if err != nil {
		return nil, fmt.Errorf("userinfo base64: %w", err)
	}

	method, password, err := splitFirst(string(decoded), ":")
	if err != nil {
		return nil, err
	}

	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("missing host")
	}

	port := 8388
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
	}

	cfg := &common.RawConfig{
		Protocol:  common.Shadowsocks,
		Name:      u.Fragment,
		Server:    host,
		Port:      port,
		Password:  password,
		Method:    method,
		Transport: common.UDP,
		Security:  "none",
	}

	logger.Debugf("ss", "SIP002 → name=%q server=%s:%d method=%s",
		cfg.Name, cfg.Server, cfg.Port, cfg.Method)
	return cfg, nil
}

func (s *ssParser) parseLegacy(uri string) (*common.RawConfig, error) {
	raw := strings.TrimPrefix(uri, "ss://")

	name := ""
	if idx := strings.Index(raw, "#"); idx != -1 {
		name, _ = url.QueryUnescape(raw[idx+1:])
		raw = raw[:idx]
	}

	decoded, err := tryBase64(raw)
	if err != nil {
		return nil, fmt.Errorf("legacy base64: %w", err)
	}
	content := string(decoded)

	// Split at the LAST '@' to separate method:password from host:port
	atIdx := strings.LastIndex(content, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("missing @ in legacy content")
	}

	method, password, err := splitFirst(content[:atIdx], ":")
	if err != nil {
		return nil, err
	}

	host, portStr, err := splitHostPort(content[atIdx+1:])
	if err != nil {
		return nil, err
	}

	port := 8388
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
	}

	cfg := &common.RawConfig{
		Protocol:  common.Shadowsocks,
		Name:      name,
		Server:    host,
		Port:      port,
		Password:  password,
		Method:    method,
		Transport: common.UDP,
		Security:  "none",
	}

	logger.Debugf("ss", "legacy → name=%q server=%s:%d method=%s",
		cfg.Name, cfg.Server, cfg.Port, cfg.Method)
	return cfg, nil
}

func splitFirst(s, sep string) (string, string, error) {
	idx := strings.Index(s, sep)
	if idx == -1 {
		return "", "", fmt.Errorf("missing %q separator in %q", sep, s)
	}
	return s[:idx], s[idx+len(sep):], nil
}

func splitHostPort(s string) (string, string, error) {
	if strings.HasPrefix(s, "[") {
		// IPv6 literal
		close := strings.Index(s, "]")
		if close == -1 {
			return "", "", fmt.Errorf("malformed IPv6 address")
		}
		host := s[1:close]
		rest := s[close+1:]
		if strings.HasPrefix(rest, ":") {
			return host, rest[1:], nil
		}
		return host, "", nil
	}
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s, "", nil
	}
	return s[:idx], s[idx+1:], nil
}
