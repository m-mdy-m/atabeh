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
// socks://  — SOCKS4/5 with optional auth (socks://, socks4://, socks5://)
//
//	https://datatracker.ietf.org/doc/html/rfc1928
//
// https://github.com/XTLS/Xray-core/discussions/5171
// https://github.com/DroidProger/XrayKeyParser
package parsers

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

const tag = "parser"

type Parser interface {
	Protocol() common.Kind
	ParseURI(uri string) (*common.RawConfig, error)
}

var registry = map[common.Kind]Parser{}

func Register(p Parser) {
	registry[p.Protocol()] = p
	logger.Debugf(tag, "registered parser: %s", p.Protocol())
}
func GetParser(proto common.Kind) Parser {
	return registry[proto]
}

func TryDecodeBase64Block(data string) ([]string, error) {
	data = strings.TrimSpace(data)

	var decoded []byte
	var err error
	for _, enc := range [](*base64.Encoding){
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.RawURLEncoding,
	} {
		decoded, err = enc.DecodeString(data)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	var uris []string
	for _, line := range strings.Split(string(decoded), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			uris = append(uris, line)
		}
	}

	logger.Infof(tag, "decoded base64 block → %d line(s)", len(uris))
	return uris, nil
}
func ParseAll(texts []string) ([]*common.RawConfig, error) {
	var configs []*common.RawConfig

	for idx, text := range texts {
		logger.Debugf(tag, "[%d] parsing: %.80s…", idx, text)

		uris := ExtractConfigs(text)
		if len(uris) == 0 {
			logger.Warnf(tag, "[%d] no URIs found, skipping", idx)
			continue
		}

		logger.Infof(tag, "[%d] extracted %d URI(s)", idx, len(uris))

		for _, uri := range uris {
			logger.Infof(tag, "[URI] %s", uri)

			proto := detectProtocol(uri)
			if proto == "" {
				logger.Warnf(tag, "unknown scheme, skipping: %.60s", uri)
				continue
			}

			parser := GetParser(proto)
			logger.Infof("[TEST-man]", "[TEST]\"%+v\"", parser)
			if parser == nil {
				logger.Warnf(tag, "no parser for protocol %s, skipping", proto)
				continue
			}

			cfg, err := parser.ParseURI(uri)
			if err != nil {
				logger.Errorf(tag, "%s parse error: %v", proto, err)
				continue
			}

			cfg.Source = "uri"
			configs = append(configs, cfg)
			logger.Infof(tag, "OK  proto=%s name=%q server=%s:%d",
				cfg.Protocol, cfg.Name, cfg.Server, cfg.Port)
		}
	}

	logger.Infof(tag, "parsed %d configs from %d texts", len(configs), len(texts))
	return configs, nil
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
	case strings.HasPrefix(uri, "socks5://"),
		strings.HasPrefix(uri, "socks4://"),
		strings.HasPrefix(uri, "socks://"):
		return common.Socks
	default:
		return ""
	}
}
