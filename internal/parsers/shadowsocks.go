package parsers

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

func init() {
	Register(&ssParser{})
}

type ssParser struct{}

func (s *ssParser) Protocol() common.Kind {
	return common.Shadowsocks
}

func (s *ssParser) ParseURI(uri string) (*common.RawConfig, error) {
	logger.Debugf("ss", "parsing URI: %.100s", uri)

	cfg, err := s.parseSIP002(uri)
	if err != nil {
		logger.Debugf("ss", "SIP002 parse failed (%v), trying legacy format", err)
		cfg, err = s.parseLegacy(uri)
		if err != nil {
			return nil, fmt.Errorf("ss parse failed (both formats): %w", err)
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
		return nil, fmt.Errorf("missing userinfo (SIP002 expects base64 userinfo)")
	}

	userinfo := u.User.Username()
	decoded, err := base64.StdEncoding.DecodeString(userinfo)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(userinfo)
		if err != nil {
			decoded, err = base64.URLEncoding.DecodeString(userinfo)
			if err != nil {
				return nil, fmt.Errorf("base64 decode of userinfo failed: %w", err)
			}
		}
	}

	method, password, err := splitMethodPassword(string(decoded))
	if err != nil {
		return nil, err
	}

	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("missing server host")
	}

	port := 8388 // SS default
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
	}

	name := u.Fragment

	cfg := &common.RawConfig{
		Protocol:  common.Shadowsocks,
		Name:      name,
		Server:    host,
		Port:      port,
		Password:  password,
		Method:    method,
		Transport: common.UDP, // SS default is UDP
		Security:  "none",
	}

	logger.Debugf("ss", "SIP002 parsed -> name=%q server=%s port=%d method=%s",
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

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(raw)
		if err != nil {
			decoded, err = base64.URLEncoding.DecodeString(raw)
			if err != nil {
				return nil, fmt.Errorf("legacy base64 decode failed: %w", err)
			}
		}
	}
	content := string(decoded)
	atIdx := strings.LastIndex(content, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("missing @ in legacy SS content")
	}

	methodPass := content[:atIdx]
	hostPort := content[atIdx+1:]

	method, password, err := splitMethodPassword(methodPass)
	if err != nil {
		return nil, err
	}

	// parse host:port
	host, portStr, err := splitHostPort(hostPort)
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

	logger.Debugf("ss", "legacy parsed -> name=%q server=%s port=%d method=%s",
		cfg.Name, cfg.Server, cfg.Port, cfg.Method)

	return cfg, nil
}

func splitMethodPassword(s string) (string, string, error) {
	idx := strings.Index(s, ":")
	if idx == -1 {
		return "", "", fmt.Errorf("missing method:password separator")
	}
	return s[:idx], s[idx+1:], nil
}

func splitHostPort(s string) (string, string, error) {
	if strings.HasPrefix(s, "[") {
		// IPv6
		closeBracket := strings.Index(s, "]")
		if closeBracket == -1 {
			return "", "", fmt.Errorf("malformed IPv6 address")
		}
		host := s[1:closeBracket]
		rest := s[closeBracket+1:]
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
