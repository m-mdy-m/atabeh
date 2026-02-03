package parsers

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	uriPattern = regexp.MustCompile(
		`(?:vless|vmess|ss|trojan[45]?)://[A-Za-z0-9+/=_\-@:.?&#%\[\]]+`,
	)
)

func Extract(text string) []string {
	text = sanitizeText(text)

	matches := uriPattern.FindAllString(text, -1)

	seen := make(map[string]struct{}, len(matches))
	out := make([]string, 0, len(matches))

	for _, m := range matches {
		uri := trimTrailing(m)
		if !isUsableURI(uri) {
			continue
		}
		if _, dup := seen[uri]; dup {
			continue
		}
		seen[uri] = struct{}{}
		out = append(out, uri)
	}

	return out
}
func sanitizeText(s string) string {
	s = strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
	).Replace(s)

	s = strings.NewReplacer(
		"\u200b", "",
		"\u200c", "",
		"\u200d", "",
		"\ufeff", "",
	).Replace(s)
	s = stripEmojis(s)
	return s
}

func stripEmojis(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if isEmoji(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func isEmoji(r rune) bool {
	switch {
	case r >= 0x1F600 && r <= 0x1F64F: // emoticons
	case r >= 0x1F300 && r <= 0x1F5FF: // symbols & pictographs
	case r >= 0x1F680 && r <= 0x1F6FF: // transport & map
	case r >= 0x2600 && r <= 0x26FF: // misc symbols
	case r >= 0x2700 && r <= 0x27BF: // dingbats
	case r >= 0xFE00 && r <= 0xFE0F: // variation selectors
	case r >= 0x1F900 && r <= 0x1F9FF: // supplemental symbols
	case r >= 0x1FA70 && r <= 0x1FAFF: // symbols extended-A
	default:
		return false
	}
	return true
}

func trimTrailing(uri string) string {

	if i := strings.IndexByte(uri, '<'); i != -1 {
		uri = uri[:i]
	}
	uri = strings.TrimRight(uri, `.,;:!?)}]"'»›*_~`+"`")
	return uri
}

func isUsableURI(uri string) bool {
	if len(uri) < 10 {
		return false
	}
	if strings.HasSuffix(uri, "...") {
		return false
	}

	idx := strings.Index(uri, "://")
	if idx == -1 || len(uri) <= idx+3 {
		return false
	}

	payload := uri[idx+3:]
	if strings.ContainsRune(payload, '@') {
		return true
	}
	return looksBase64(payload)
}

func looksBase64(s string) bool {
	if len(s) < 8 {
		return false
	}
	good := 0
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '+' || r == '/' || r == '=' || r == '-' || r == '_' {
			good++
		}
	}
	return float64(good)/float64(len(s)) > 0.7
}
