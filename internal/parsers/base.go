// vless://  — XTLS/Xray-core standard
//
//	https://github.com/XTLS/Xray-core/discussions/5171
//
// vmess://  — V2Ray base64-JSON payload
//
//	https://github.com/2fly4info/V2RayNG/blob/master/README.md
//
// ss://     — Shadowsocks SIP002 + legacy base64
//
//	https://shadowsocks.org/en/config/quick-guide.html
//
// trojan:// — password-based, TLS by convention
//
//	https://trojan-gfw.github.io/trojan/protocol.html
//
//	https://datatracker.ietf.org/doc/html/rfc1928
//
// https://github.com/XTLS/Xray-core/discussions/5171
// https://github.com/DroidProger/XrayKeyParser
package parsers

import (
	"strconv"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

type Parser interface {
	Protocol() common.Kind
	ParseURI(uri string) (*common.RawConfig, error)
}

var registry = map[common.Kind]Parser{}

func Register(p Parser)                  { registry[p.Protocol()] = p }
func GetParser(proto common.Kind) Parser { return registry[proto] }

func ParseURIs(uris []string) ([]*common.RawConfig, error) {
	var configs []*common.RawConfig

	for _, uri := range uris {
		proto := detectProtocol(uri)
		if proto == "" {
			logger.Warn("parser", "unknown scheme, skipping: "+truncate(uri, 60))
			continue
		}

		p := registry[proto]
		if p == nil {
			logger.Warn("parser", "no parser registered for "+string(proto))
			continue
		}

		cfg, err := p.ParseURI(uri)
		if err != nil {
			logger.Warn("parser", string(proto)+" parse error: "+err.Error())
			continue
		}
		if cfg.Extra == nil {
			cfg.Extra = map[string]string{}
		}
		cfg.Extra["raw_uri"] = uri
		cfg.Source = "uri"
		configs = append(configs, cfg)
	}

	logger.Debug("parser", "parsed "+strconv.Itoa(len(configs))+" configs from "+strconv.Itoa(len(uris))+" URIs")
	return configs, nil
}

func ParseText(text string) ([]*common.RawConfig, error) {
	uris := Extract(text)
	if len(uris) == 0 {
		return nil, nil
	}
	return ParseURIs(uris)
}

func detectProtocol(uri string) common.Kind {
	switch {
	case strings.HasPrefix(uri, "vless://"):
		return common.Vless
	case strings.HasPrefix(uri, "vmess://"):
		return common.VMess
	case strings.HasPrefix(uri, "ss://"):
		return common.Shadowsocks
	case strings.HasPrefix(uri, "trojan://"):
		return common.Trojan
	default:
		return ""
	}
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
