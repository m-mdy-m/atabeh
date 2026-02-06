package exporter

import (
	"encoding/json"
	"fmt"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/storage"
)

type SingBoxConfig struct {
	Outbounds []SingBoxOutbound `json:"outbounds"`
	Route     SingBoxRoute      `json:"route"`
}

type SingBoxOutbound struct {
	Type     string      `json:"type"`
	Tag      string      `json:"tag"`
	Server   string      `json:"server"`
	Port     int         `json:"server_port"`
	UUID     string      `json:"uuid,omitempty"`
	Password string      `json:"password,omitempty"`
	Method   string      `json:"method,omitempty"`
	Network  string      `json:"network,omitempty"`
	TLS      *SingBoxTLS `json:"tls,omitempty"`
}

type SingBoxTLS struct {
	Enabled    bool   `json:"enabled"`
	ServerName string `json:"server_name,omitempty"`
	Insecure   bool   `json:"insecure"`
}

type SingBoxRoute struct {
	Rules []SingBoxRule `json:"rules"`
}

type SingBoxRule struct {
	Outbound string `json:"outbound"`
}

func ToSingBox(configs []*storage.ConfigRow) ([]byte, error) {
	outbounds := make([]SingBoxOutbound, 0, len(configs)+1)

	for i, cfg := range configs {
		tag := fmt.Sprintf("proxy-%d", i+1)
		out := configToOutbound(cfg, tag)
		if out != nil {
			outbounds = append(outbounds, *out)
		}
	}

	if len(outbounds) == 0 {
		return nil, fmt.Errorf("no valid outbounds generated")
	}

	outbounds = append(outbounds, SingBoxOutbound{
		Type: "direct",
		Tag:  "direct",
	})

	config := SingBoxConfig{
		Outbounds: outbounds,
		Route: SingBoxRoute{
			Rules: []SingBoxRule{
				{Outbound: "proxy-1"},
			},
		},
	}

	return json.MarshalIndent(config, "", "  ")
}

func configToOutbound(cfg *storage.ConfigRow, tag string) *SingBoxOutbound {
	out := &SingBoxOutbound{
		Tag:    tag,
		Server: cfg.Server,
		Port:   cfg.Port,
	}

	hasTLS := cfg.Security == "tls" || cfg.Security == "reality"

	switch cfg.Protocol {
	case common.Vless:
		out.Type = "vless"
		out.UUID = cfg.UUID
		out.Network = string(cfg.Transport)

	case common.VMess:
		out.Type = "vmess"
		out.UUID = cfg.UUID
		out.Network = string(cfg.Transport)

	case common.Shadowsocks:
		out.Type = "shadowsocks"
		out.Method = cfg.Method
		out.Password = cfg.Password

	case common.Trojan:
		out.Type = "trojan"
		out.Password = cfg.Password
		hasTLS = true

	default:
		return nil
	}

	if hasTLS {
		out.TLS = &SingBoxTLS{
			Enabled:    true,
			ServerName: cfg.Server,
			Insecure:   true,
		}

		if cfg.Extra != "" {

		}
	}

	return out
}
