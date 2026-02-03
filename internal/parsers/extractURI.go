package parsers

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

const extractTag = "extract"

var (
	vlessPattern     = regexp.MustCompile(`\bvless://[A-Za-z0-9+/=_\-@:\.?&#%]+`)
	vmessPattern     = regexp.MustCompile(`\bvmess://[A-Za-z0-9+/=_\-]+`)
	ssPattern        = regexp.MustCompile(`\bss://[A-Za-z0-9+/=_\-@:\.?&#%]+`)
	trojanPattern    = regexp.MustCompile(`\btrojan://[A-Za-z0-9+/=_\-@:\.?&#%]+`)
	socksPattern     = regexp.MustCompile(`\bsocks[45]?://[A-Za-z0-9+/=_\-@:\.?&#%]+`)
	protocolPatterns = map[common.Kind]*regexp.Regexp{
		common.Vless:       vlessPattern,
		common.VMess:       vmessPattern,
		common.Shadowsocks: ssPattern,
		common.Trojan:      trojanPattern,
		common.Socks:       socksPattern,
	}
)

type URIExtractor struct {
	cleanPunctuation bool
	deduplicateURIs  bool
}

func NewURIExtractor() *URIExtractor {
	return &URIExtractor{
		cleanPunctuation: true,
		deduplicateURIs:  true,
	}
}

func ExtractURIs(text string) []string {
	extractor := NewURIExtractor()
	return extractor.Extract(text)
}
func (e *URIExtractor) Extract(text string) []string {
	logger.Debugf(extractTag, "extracting URIs from text (%d chars)", len(text))
	cleanedText := e.cleanText(text)

	seen := make(map[string]bool)
	var results []string

	for protocol, pattern := range protocolPatterns {
		matches := pattern.FindAllString(cleanedText, -1)
		logger.Debugf(extractTag, "found %d potential %s URIs", len(matches), protocol)

		for _, match := range matches {
			uri := e.cleanURI(match)
			if !e.isValidURI(uri) {
				logger.Debugf(extractTag, "invalid URI filtered: %s", uri[:min(len(uri), 50)])
				continue
			}
			if e.deduplicateURIs {
				if seen[uri] {
					continue
				}
				seen[uri] = true
			}

			results = append(results, uri)
		}
	}

	logger.Infof(extractTag, "extracted %d unique URI(s) from text", len(results))
	return results
}

func (e *URIExtractor) cleanText(text string) string {
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "\u200b", "")
	text = strings.ReplaceAll(text, "\u200c", "")
	text = strings.ReplaceAll(text, "\u200d", "")
	text = strings.ReplaceAll(text, "\ufeff", "")

	return text
}
func (e *URIExtractor) cleanURI(uri string) string {
	if !e.cleanPunctuation {
		return uri
	}
	trailingChars := ".,;:!?)}\"]'»›"
	uri = strings.TrimRight(uri, trailingChars)
	uri = strings.TrimSpace(uri)
	if idx := strings.Index(uri, "<"); idx != -1 {
		uri = uri[:idx]
	}
	uri = strings.TrimRight(uri, "*_~`")

	return uri
}
func (e *URIExtractor) isValidURI(uri string) bool {
	if len(uri) < 10 {
		return false
	}
	hasProtocol := false
	protocols := []string{"vless://", "vmess://", "ss://", "trojan://", "socks://", "socks4://", "socks5://"}
	for _, proto := range protocols {
		if strings.HasPrefix(uri, proto) {
			hasProtocol = true
			break
		}
	}
	if !hasProtocol {
		return false
	}
	hasAtOrBase64 := strings.Contains(uri, "@") || containsBase64Chars(uri)
	if !hasAtOrBase64 {
		return false
	}
	if strings.HasSuffix(uri, "...") {
		return false
	}

	return true
}

func containsBase64Chars(s string) bool {
	base64Count := 0
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '+' || r == '/' || r == '=' {
			base64Count++
		}
	}
	return float64(base64Count)/float64(len(s)) > 0.5
}

func ExtractConfigs(text string) []string {
	text = removeEmojis(text)
	extractor := NewURIExtractor()
	return extractor.Extract(text)
}

func removeEmojis(text string) string {
	var builder strings.Builder
	for _, r := range text {
		// Skip emoji ranges
		if (r >= 0x1F600 && r <= 0x1F64F) ||
			(r >= 0x1F300 && r <= 0x1F5FF) ||
			(r >= 0x1F680 && r <= 0x1F6FF) ||
			(r >= 0x2600 && r <= 0x26FF) ||
			(r >= 0x2700 && r <= 0x27BF) ||
			(r >= 0xFE00 && r <= 0xFE0F) ||
			(r >= 0x1F900 && r <= 0x1F9FF) ||
			(r >= 0x1FA70 && r <= 0x1FAFF) {
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}
