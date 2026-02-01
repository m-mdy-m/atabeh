package parsers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

func init() {
	Register(&vmessParser{})
}

type vmessParser struct{}

func (v *vmessParser) Protocol() common.Kind {
	return common.VMess
}

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

// ParseURI parses: vmess://BASE64PAYLOAD
func (v *vmessParser) ParseURI(uri string) (*common.RawConfig, error) {
	logger.Debugf("vmess", "parsing URI: %.80s...", uri)

	raw := strings.TrimPrefix(uri, "vmess://")
	raw = strings.TrimSpace(raw)

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(raw)
		if err != nil {
			decoded, err = base64.RawStdEncoding.DecodeString(raw)
			if err != nil {
				return nil, fmt.Errorf("vmess base64 decode failed: %w", err)
			}
		}
	}

	logger.Debugf("vmess", "decoded payload: %s", string(decoded))

	var payload vmessJSON
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, fmt.Errorf("vmess JSON parse failed: %w", err)
	}

	if payload.Server == "" {
		return nil, fmt.Errorf("missing server address in vmess config")
	}

	port, err := parsePort(payload.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid vmess port: %w", err)
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

	cfg := &common.RawConfig{
		Protocol:  common.VMess,
		Name:      payload.Name,
		Server:    payload.Server,
		Port:      port,
		UUID:      payload.UUID,
		Transport: transport,
		Security:  security,
		Extra:     extra,
	}

	logger.Debugf("vmess", "parsed -> name=%q server=%s port=%d transport=%s security=%s",
		cfg.Name, cfg.Server, cfg.Port, cfg.Transport, cfg.Security)

	return cfg, nil
}

func parsePort(raw any) (int, error) {
	switch v := raw.(type) {
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	case nil:
		return 443, nil // default
	default:
		return 0, fmt.Errorf("unexpected port type: %T", raw)
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
