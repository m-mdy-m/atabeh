package parsers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

func init() { Register(&vmessParser{}) }

type vmessParser struct{}

func (v *vmessParser) Protocol() common.Kind { return common.VMess }

type vmessJSON struct {
	Name     string `json:"ps"`
	Version  string `json:"ver"`
	UUID     string `json:"id"`
	AltID    string `json:"aid"`
	Security string `json:"scy"`
	Server   string `json:"add"`
	Port     any    `json:"port"`
	Network  string `json:"net"`
	Type     string `json:"type"`
	TLS      string `json:"tls"`
	Path     string `json:"path"`
	Host     string `json:"host"`
}

// ParseURI parses: vmess://BASE64(json)
func (v *vmessParser) ParseURI(uri string) (*common.RawConfig, error) {
	logger.Debugf("vmess", "parsing: %.80s…", uri)

	raw := strings.TrimSpace(strings.TrimPrefix(uri, "vmess://"))

	decoded, err := tryBase64(raw)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	var payload vmessJSON
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, fmt.Errorf("JSON unmarshal: %w", err)
	}

	if payload.Server == "" {
		return nil, fmt.Errorf("missing server (add)")
	}

	port, err := flexPort(payload.Port)
	if err != nil {
		return nil, fmt.Errorf("port: %w", err)
	}

	transport := detectTransport(payload.Network)

	security := payload.TLS
	if security == "" {
		security = "none"
	}

	extra := map[string]string{}
	if payload.AltID != "" && payload.AltID != "0" {
		extra["aid"] = payload.AltID
	}
	if payload.Type != "" && payload.Type != "none" {
		extra["camouflage"] = payload.Type
	}
	if payload.Path != "" {
		extra["path"] = payload.Path
	}
	if payload.Host != "" {
		extra["host"] = payload.Host
	}
	if payload.Security != "" {
		extra["encryption"] = payload.Security
	}

	name, err := url.QueryUnescape(payload.Name)
	if err != nil {
		name = payload.Name
	}
	cfg := &common.RawConfig{
		Protocol:  common.VMess,
		Name:      name,
		Server:    payload.Server,
		Port:      port,
		UUID:      payload.UUID,
		Transport: transport,
		Security:  security,
		Extra:     extra,
	}

	logger.Debugf("vmess", "→ name=%q server=%s:%d transport=%s security=%s",
		cfg.Name, cfg.Server, cfg.Port, cfg.Transport, cfg.Security)
	return cfg, nil
}

func tryBase64(s string) ([]byte, error) {
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.RawURLEncoding,
	} {
		if b, err := enc.DecodeString(s); err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf("all base64 variants failed")
}

func flexPort(raw any) (int, error) {
	switch v := raw.(type) {
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	case nil:
		return 443, nil
	default:
		return 0, fmt.Errorf("unexpected type %T", raw)
	}
}

func detectTransport(network string) common.Kind {
	switch strings.ToLower(network) {
	case "ws":
		return common.WS
	case "h2":
		return common.H2
	case "grpc":
		return common.GRPC
	case "udp", "kcp":
		return common.UDP
	default:
		return common.TCP
	}
}
