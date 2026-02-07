package exporter

import (
	"encoding/json"
	"fmt"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/storage"
)

func ToXray(configs []*storage.ConfigRow, enableStats bool) ([]byte, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("no configs")
	}

	config := map[string]interface{}{
		"log": map[string]interface{}{
			"loglevel": "warning",
		},
		"inbounds": []map[string]interface{}{
			{
				"tag":      "socks-in",
				"port":     10808,
				"listen":   "127.0.0.1",
				"protocol": "socks",
				"settings": map[string]interface{}{
					"udp": true,
				},
			},
			{
				"tag":      "http-in",
				"port":     10809,
				"listen":   "127.0.0.1",
				"protocol": "http",
				"settings": map[string]interface{}{},
			},
		},
		"outbounds": []map[string]interface{}{
			buildOutbound(configs[0]),
			{
				"tag":      "direct",
				"protocol": "freedom",
				"settings": map[string]interface{}{},
			},
			{
				"tag":      "block",
				"protocol": "blackhole",
				"settings": map[string]interface{}{},
			},
		},
		"routing": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"type":        "field",
					"inboundTag":  []string{"socks-in", "http-in"},
					"outboundTag": "proxy",
				},
			},
		},
	}

	if enableStats {
		config["stats"] = map[string]interface{}{}
		config["api"] = map[string]interface{}{
			"tag":      "api",
			"services": []string{"StatsService"},
		}
		config["policy"] = map[string]interface{}{
			"system": map[string]interface{}{
				"statsInboundUplink":    true,
				"statsInboundDownlink":  true,
				"statsOutboundUplink":   true,
				"statsOutboundDownlink": true,
			},
		}

		inbounds := config["inbounds"].([]map[string]interface{})
		inbounds = append(inbounds, map[string]interface{}{
			"tag":      "api",
			"port":     10085,
			"listen":   "127.0.0.1",
			"protocol": "dokodemo-door",
			"settings": map[string]interface{}{
				"address": "127.0.0.1",
			},
		})
		config["inbounds"] = inbounds
	}

	return json.MarshalIndent(config, "", "  ")
}

func buildOutbound(cfg *storage.ConfigRow) map[string]interface{} {
	outbound := map[string]interface{}{
		"tag":      "proxy",
		"protocol": string(cfg.Protocol),
	}

	switch cfg.Protocol {
	case common.Vless:
		outbound["settings"] = map[string]interface{}{
			"vnext": []map[string]interface{}{
				{
					"address": cfg.Server,
					"port":    cfg.Port,
					"users": []map[string]interface{}{
						{
							"id":         cfg.UUID,
							"encryption": "none",
						},
					},
				},
			},
		}

	case common.VMess:
		outbound["settings"] = map[string]interface{}{
			"vnext": []map[string]interface{}{
				{
					"address": cfg.Server,
					"port":    cfg.Port,
					"users": []map[string]interface{}{
						{
							"id":       cfg.UUID,
							"alterId":  0,
							"security": "auto",
						},
					},
				},
			},
		}

	case common.Trojan:
		outbound["settings"] = map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"address":  cfg.Server,
					"port":     cfg.Port,
					"password": cfg.Password,
				},
			},
		}

	case common.Shadowsocks:
		outbound["settings"] = map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"address":  cfg.Server,
					"port":     cfg.Port,
					"method":   cfg.Method,
					"password": cfg.Password,
				},
			},
		}
	}

	streamSettings := map[string]interface{}{
		"network": string(cfg.Transport),
	}

	if cfg.Security == "tls" {
		streamSettings["security"] = "tls"
		streamSettings["tlsSettings"] = map[string]interface{}{
			"allowInsecure": true,
			"serverName":    cfg.Server,
		}
	}

	switch cfg.Transport {
	case common.WS:
		streamSettings["wsSettings"] = map[string]interface{}{
			"path": "/",
		}

		if cfg.Extra != "" {
			var extra map[string]string
			if err := json.Unmarshal([]byte(cfg.Extra), &extra); err == nil {
				if path, ok := extra["path"]; ok && path != "" {
					streamSettings["wsSettings"].(map[string]interface{})["path"] = path
				}
				if host, ok := extra["host"]; ok && host != "" {
					streamSettings["wsSettings"].(map[string]interface{})["headers"] = map[string]string{
						"Host": host,
					}
				}
			}
		}

	case common.GRPC:
		streamSettings["grpcSettings"] = map[string]interface{}{
			"serviceName": "",
		}

		if cfg.Extra != "" {
			var extra map[string]string
			if err := json.Unmarshal([]byte(cfg.Extra), &extra); err == nil {
				if svc, ok := extra["serviceName"]; ok && svc != "" {
					streamSettings["grpcSettings"].(map[string]interface{})["serviceName"] = svc
				}
			}
		}

	case common.H2:
		streamSettings["httpSettings"] = map[string]interface{}{
			"path": "/",
		}

		if cfg.Extra != "" {
			var extra map[string]string
			if err := json.Unmarshal([]byte(cfg.Extra), &extra); err == nil {
				if path, ok := extra["path"]; ok && path != "" {
					streamSettings["httpSettings"].(map[string]interface{})["path"] = path
				}
				if host, ok := extra["host"]; ok && host != "" {
					streamSettings["httpSettings"].(map[string]interface{})["host"] = []string{host}
				}
			}
		}
	}

	outbound["streamSettings"] = streamSettings

	return outbound
}
