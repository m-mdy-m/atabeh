// https://github.com/XTLS/Xray-core/discussions/5171
// https://github.com/DroidProger/XrayKeyParser
package parsers

import (
	"encoding/base64"
	"fmt"
	"regexp"
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
	logger.Debugf(tag, "registered parser for protocol: %s", p.Protocol())
}
func GetParser(proto common.Kind) Parser {
	return registry[proto]
}

var (
	vlessPattern = regexp.MustCompile(`vless://[^\s\r\n]+`)
	vmessPattern = regexp.MustCompile(`vmess://[^\s\r\n]+`)
	ssPattern    = regexp.MustCompile(`ss://[^\s\r\n]+`)
)

var allPatterns = []*regexp.Regexp{vlessPattern, vmessPattern, ssPattern}

func ExtractURIs(text string) []string {
	logger.Debugf(tag, "extracting URIs from text (%d chars)", len(text))

	seen := map[string]bool{}
	var results []string

	for _, pattern := range allPatterns {
		matches := pattern.FindAllString(text, -1)
		for _, m := range matches {
			m = cleanURI(m)
			if seen[m] {
				continue
			}
			seen[m] = true
			results = append(results, m)
		}
	}

	logger.Infof(tag, "extracted %d unique URI(s) from text", len(results))
	return results
}

func cleanURI(uri string) string {
	trimChars := ".,;:!?)}\"]'»›"
	uri = strings.TrimRight(uri, trimChars)
	return uri
}

func TryDecodeBase64Block(data string) ([]string, error) {
	data = strings.TrimSpace(data)
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(data)
		if err != nil {
			decoded, err = base64.RawStdEncoding.DecodeString(data)
			if err != nil {
				decoded, err = base64.RawURLEncoding.DecodeString(data)
				if err != nil {
					return nil, fmt.Errorf("base64 decode failed: %w", err)
				}
			}
		}
	}

	lines := strings.Split(strings.TrimSpace(string(decoded)), "\n")
	var uris []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		uris = append(uris, line)
	}

	logger.Infof(tag, "decoded base64 block -> %d line(s)", len(uris))
	return uris, nil
}

func ParseAll(uris []string) ([]*common.RawConfig, error) {
	var configs []*common.RawConfig

	for i, uri := range uris {
		logger.Debugf(tag, "[%d] attempting to parse: %.80s...", i, uri)

		proto := detectProtocol(uri)
		if proto == "" {
			logger.Warnf(tag, "[%d] unknown protocol, skipping: %.60s", i, uri)
			continue
		}

		p := GetParser(proto)
		if p == nil {
			logger.Warnf(tag, "[%d] no parser registered for %s", i, proto)
			continue
		}

		cfg, err := p.ParseURI(uri)
		if err != nil {
			logger.Errorf(tag, "[%d] parse error (%s): %v", i, proto, err)
			continue
		}

		cfg.Source = "uri"
		configs = append(configs, cfg)
		logger.Infof(tag, "[%d] parsed OK: protocol=%s name=%q server=%s:%d", i, cfg.Protocol, cfg.Name, cfg.Server, cfg.Port)
	}

	logger.Infof(tag, "parsed %d/%d configs successfully", len(configs), len(uris))
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
	default:
		return ""
	}
}
