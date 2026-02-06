package normalizer

import (
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

var (
	symbolPattern = regexp.MustCompile(`[«»‹›「」【】〔〕（）()[\]{}⟨⟩]+`)
	emojiPattern  = regexp.MustCompile(
		`[` +
			`\x{1F1E6}-\x{1F1FF}` +
			`\x{1F300}-\x{1F5FF}` +
			`\x{1F600}-\x{1F64F}` +
			`\x{1F680}-\x{1F6FF}` +
			`\x{1F900}-\x{1F9FF}` +
			`\x{1FA70}-\x{1FAFF}` +
			`\x{2600}-\x{26FF}` +
			`\x{2700}-\x{27BF}` +
			`]+`,
	)
	locationPattern = regexp.MustCompile(`^[🇦-🇿]{2}\s*\d+\s*[-–—]\s*|^\d+\s*[-–—]\s*|^[A-Z]{2}[-–—]\d+\s*[-–—]\s*`)
	repeatedDash    = regexp.MustCompile(`[-–—\s]{2,}`)
)

func CleanName(name string) string {
	if name == "" {
		return ""
	}

	if decoded, err := url.QueryUnescape(name); err == nil {
		name = decoded
	}

	name = symbolPattern.ReplaceAllString(name, " ")
	name = emojiPattern.ReplaceAllString(name, "")

	name = locationPattern.ReplaceAllString(name, "")
	name = strings.Map(func(r rune) rune {
		if r == '\uFE0F' || r == '\uFE0E' || r == '\u200D' ||
			(r >= 0xE0100 && r <= 0xE01EF) {
			return -1
		}
		if unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, name)

	name = strings.TrimLeftFunc(name, func(r rune) bool {
		return unicode.IsPunct(r) || unicode.IsSymbol(r)
	})

	name = repeatedDash.ReplaceAllString(name, " ")

	name = strings.TrimSpace(name)
	name = strings.Join(strings.Fields(name), " ")

	return name
}

func ExtractProfileName(source string) string {
	source = strings.TrimSpace(source)

	if strings.HasPrefix(source, "http") {
		if idx := strings.LastIndex(source, "#"); idx != -1 && idx < len(source)-1 {
			fragment := source[idx+1:]
			cleaned := CleanName(fragment)
			if cleaned != "" {
				return cleaned
			}
		}

		parts := strings.Split(source, "/")
		for i := len(parts) - 1; i >= 0; i-- {
			part := parts[i]
			if part == "" || part == "raw" || part == "main" {
				continue
			}

			for _, ext := range []string{".txt", ".conf", ".config", ".json"} {
				part = strings.TrimSuffix(part, ext)
			}

			cleaned := CleanName(part)
			if cleaned != "" && len(cleaned) > 2 {
				return titleCase(cleaned)
			}
		}

		if strings.Contains(source, "://") {
			domain := strings.Split(strings.Split(source, "://")[1], "/")[0]
			parts := strings.Split(domain, ".")
			if len(parts) >= 2 {
				return titleCase(CleanName(parts[0]))
			}
		}
	}

	if strings.Contains(source, "://") {
		if idx := strings.LastIndex(source, "#"); idx != -1 && idx < len(source)-1 {
			fragment := source[idx+1:]
			cleaned := CleanName(fragment)
			if cleaned != "" {
				return cleaned
			}
		}
	}

	return "Configs"
}

func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
