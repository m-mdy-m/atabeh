package parsers

import (
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

type MixedParseResult struct {
	Subscriptions []string
	DirectConfigs []*common.RawConfig
}

func ParseMixedContent(text string) (*MixedParseResult, error) {
	result := &MixedParseResult{
		Subscriptions: []string{},
		DirectConfigs: []*common.RawConfig{},
	}

	subs := extractSubscriptionURLs(text)
	result.Subscriptions = append(result.Subscriptions, subs...)

	uris := Extract(text)
	if len(uris) > 0 {
		configs, err := ParseURIs(uris)
		if err != nil {
			logger.Warnf("parser", "parse direct configs: %v", err)
		} else {
			result.DirectConfigs = append(result.DirectConfigs, configs...)
		}
	}

	logger.Infof("parser", "mixed content: %d subscriptions, %d direct configs",
		len(result.Subscriptions), len(result.DirectConfigs))

	return result, nil
}

func extractSubscriptionURLs(text string) []string {
	var subs []string
	seen := make(map[string]bool)

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if IsSubscriptionURL(line) {
			if !seen[line] {
				subs = append(subs, line)
				seen[line] = true
			}
		}
	}

	return subs
}

func IsSubscriptionURL(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))

	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return false
	}

	if strings.Contains(s, "vless://") || strings.Contains(s, "vmess://") ||
		strings.Contains(s, "ss://") || strings.Contains(s, "trojan://") {
		return false
	}

	indicators := []string{
		"raw.githubusercontent.com",
		"gist.githubusercontent.com",
		"/sub", "/subscription", "/config",
		".txt", "/raw/",
	}

	for _, indicator := range indicators {
		if strings.Contains(s, indicator) {
			return true
		}
	}

	return false
}

func FetchAndParseAll(source string) ([]*common.RawConfig, error) {
	logger.Infof("parser", "fetching: %s", truncSource(source))

	if !strings.HasPrefix(source, "http") {
		return parseRawText(source)
	}

	text, err := fetchWithRetry(source)
	if err != nil {
		return nil, err
	}

	text = tryDecodeWholeBody(text)

	mixed, err := ParseMixedContent(text)
	if err != nil {
		logger.Warnf("parser", "parse mixed: %v", err)
	}

	allConfigs := make([]*common.RawConfig, 0)
	allConfigs = append(allConfigs, mixed.DirectConfigs...)

	for _, subURL := range mixed.Subscriptions {
		logger.Infof("parser", "fetching nested: %s", subURL)
		subText, err := fetchWithRetry(subURL)
		if err != nil {
			logger.Warnf("parser", "fetch nested %s: %v", subURL, err)
			continue
		}

		subText = tryDecodeWholeBody(subText)
		subURIs := Extract(subText)

		if len(subURIs) > 0 {
			subConfigs, err := ParseURIs(subURIs)
			if err != nil {
				logger.Warnf("parser", "parse nested %s: %v", subURL, err)
				continue
			}
			allConfigs = append(allConfigs, subConfigs...)
		}
	}

	logger.Infof("parser", "extracted %d total configs", len(allConfigs))
	return allConfigs, nil
}

func parseRawText(text string) ([]*common.RawConfig, error) {
	uris := Extract(text)
	if len(uris) == 0 {
		return nil, nil
	}
	return ParseURIs(uris)
}

func truncSource(s string) string {
	if len(s) > 60 {
		return s[:57] + "..."
	}
	return s
}
