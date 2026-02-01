package parsers

import (
	"github.com/m-mdy-m/atabeh/internal/logger"
)

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
