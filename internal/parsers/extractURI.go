package parsers

import (
	"regexp"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/logger"
)

var (
	vlessPattern  = regexp.MustCompile(`(?:^|\s)vless://[^\s\r\n]+`)
	vmessPattern  = regexp.MustCompile(`(?:^|\s)vmess://[^\s\r\n]+`)
	ssPattern     = regexp.MustCompile(`(?:^|\s)ss://[^\s\r\n]+`)
	trojanPattern = regexp.MustCompile(`(?:^|\s)trojan://[^\s\r\n]+`)
	socksPattern  = regexp.MustCompile(`(?:^|\s)socks(?:4|5)?://[^\s\r\n]+`)
)

var allPatterns = []*regexp.Regexp{
	vlessPattern, vmessPattern, ssPattern, trojanPattern, socksPattern,
}

func cleanURI(uri string) string {
	uri = strings.TrimSpace(uri)
	// Iteratively strip trailing punctuation — some URIs end with ")." etc.
	for {
		trimmed := strings.TrimRight(uri, ".,;:!?)}\"]'»›⟩〕】」›")
		if trimmed == uri {
			break
		}
		uri = trimmed
	}
	return uri
}

func normaliseEntities(text string) string {
	r := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&#38;", "&",
	)
	return r.Replace(text)
}
func ExtractURIs(text string) []string {
	text = normaliseEntities(text)
	logger.Debugf(tag, "extracting URIs from text (%d chars)", len(text))

	seen := map[string]bool{}
	var results []string

	for _, pattern := range allPatterns {
		matches := pattern.FindAllString(text, -1)
		for _, m := range matches {
			m = cleanURI(m)
			if m == "" || seen[m] {
				continue
			}
			seen[m] = true
			results = append(results, m)
		}
	}

	logger.Infof(tag, "extracted %d unique URI(s)", len(results))
	return results
}
