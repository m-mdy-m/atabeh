package parsers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
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

// ParseURI parses  vmess://BASE64(json)
func (v *vmessParser) ParseURI(uri string) (*common.RawConfig, error) {
	raw := strings.TrimSpace(strings.TrimPrefix(uri, "vmess://"))

	decoded, err := tryBase64(raw)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	var p vmessJSON
	if err := json.Unmarshal(decoded, &p); err != nil {
		return nil, fmt.Errorf("JSON unmarshal: %w", err)
	}
	if p.Server == "" {
		return nil, fmt.Errorf("missing server (add)")
	}

	port, err := flexPort(p.Port)
	if err != nil {
		return nil, fmt.Errorf("port: %w", err)
	}

	security := p.TLS
	if security == "" {
		security = "none"
	}

	extra := map[string]string{}
	if p.AltID != "" && p.AltID != "0" {
		extra["aid"] = p.AltID
	}
	if p.Type != "" && p.Type != "none" {
		extra["camouflage"] = p.Type
	}
	if p.Path != "" {
		extra["path"] = p.Path
	}
	if p.Host != "" {
		extra["host"] = p.Host
	}
	if p.Security != "" {
		extra["encryption"] = p.Security
	}

	return &common.RawConfig{
		Protocol:  common.VMess,
		Name:      decodeName(p.Name),
		Server:    p.Server,
		Port:      port,
		UUID:      p.UUID,
		Transport: vmessTransport(p.Network),
		Security:  security,
		Extra:     extra,
	}, nil
}

func vmessTransport(network string) common.Kind {
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
