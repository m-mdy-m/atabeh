package parsers

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
)

func init() { Register(&ssParser{}) }

type ssParser struct{}

func (s *ssParser) Protocol() common.Kind { return common.Shadowsocks }

// ParseURI handles both formats:
//
//	SIP002:  ss://base64(method:password)@host:port#name
//	Legacy:  ss://base64(method:password@host:port)#name
func (s *ssParser) ParseURI(uri string) (*common.RawConfig, error) {
	cfg, err := s.parseSIP002(uri)
	if err == nil {
		return cfg, nil
	}
	// SIP002 failed — try legacy
	cfg2, err2 := s.parseLegacy(uri)
	if err2 != nil {
		return nil, fmt.Errorf("SIP002: %w; legacy: %w", err, err2)
	}
	return cfg2, nil
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

	return &common.RawConfig{
		Protocol:  common.Shadowsocks,
		Name:      decodeName(u.Fragment),
		Server:    host,
		Port:      port,
		Password:  password,
		Method:    method,
		Transport: common.UDP,
		Security:  "none",
	}, nil
}

func (s *ssParser) parseLegacy(uri string) (*common.RawConfig, error) {
	raw := strings.TrimPrefix(uri, "ss://")

	// fragment (name) is after '#'
	name := ""
	if idx := strings.IndexByte(raw, '#'); idx != -1 {
		name = decodeName(raw[idx+1:])
		raw = raw[:idx]
	}

	decoded, err := tryBase64(raw)
	if err != nil {
		return nil, fmt.Errorf("legacy base64: %w", err)
	}
	content := string(decoded)

	// split at LAST '@' — password itself may contain '@'
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

	return &common.RawConfig{
		Protocol:  common.Shadowsocks,
		Name:      name,
		Server:    host,
		Port:      port,
		Password:  password,
		Method:    method,
		Transport: common.UDP,
		Security:  "none",
	}, nil
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
		close := strings.IndexByte(s, ']')
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
	idx := strings.LastIndexByte(s, ':')
	if idx == -1 {
		return s, "", nil
	}
	return s[:idx], s[idx+1:], nil
}
