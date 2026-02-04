package parsers

import (
	"fmt"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

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

type MixedParseResult struct {
	Subscriptions []string
	DirectConfigs []*common.RawConfig
}

func extractSubscriptionURLs(text string) []string {
	var subs []string
	seen := make(map[string]bool)

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if isSubscriptionURL(line) {
			if !seen[line] {
				subs = append(subs, line)
				seen[line] = true
			}
		}
	}

	return subs
}

func isSubscriptionURL(s string) bool {
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

func ExtractProfileName(source string) string {

	source = strings.TrimSpace(source)

	if strings.HasPrefix(source, "http") {

		if idx := strings.LastIndex(source, "#"); idx != -1 && idx < len(source)-1 {
			name := source[idx+1:]
			if name != "" {
				return cleanProfileName(name)
			}
		}

		parts := strings.Split(source, "/")
		for i := len(parts) - 1; i >= 0; i-- {
			part := parts[i]
			if part != "" && part != "raw" && part != "main" && !strings.HasPrefix(part, "http") {

				name := strings.TrimSuffix(part, ".txt")
				name = strings.TrimSuffix(name, ".conf")
				name = strings.TrimSuffix(name, ".config")
				if name != "" {
					return cleanProfileName(name)
				}
			}
		}

		if strings.Contains(source, "://") {
			domainPart := strings.Split(strings.Split(source, "://")[1], "/")[0]
			parts := strings.Split(domainPart, ".")
			if len(parts) >= 2 {
				return cleanProfileName(parts[0])
			}
		}
	}

	if strings.Contains(source, "://") {

		if idx := strings.LastIndex(source, "#"); idx != -1 && idx < len(source)-1 {
			name := source[idx+1:]
			if name != "" {
				return cleanProfileName(name)
			}
		}
	}

	return "Unknown Profile"
}

func cleanProfileName(name string) string {
	name = strings.TrimSpace(name)

	name = strings.ReplaceAll(name, "%20", " ")
	name = strings.ReplaceAll(name, "%D8%B3", "س")
	name = strings.ReplaceAll(name, "%D8%B1", "ر")

	prefixes := []string{"@", "subscription_", "config_", "proxy_"}
	for _, prefix := range prefixes {
		name = strings.TrimPrefix(name, prefix)
	}

	if len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}

	return name
}

func FetchAndParseAll(source string) ([]*common.RawConfig, error) {
	logger.Infof("parser", "fetching and parsing: %s", source)

	if !strings.HasPrefix(source, "http") {
		return ParseText(source)
	}

	text, err := fetchWithRetry(source)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	text = tryDecodeWholeBody(text)

	mixed, err := ParseMixedContent(text)
	if err != nil {
		logger.Warnf("parser", "parse mixed content: %v", err)
	}

	allConfigs := make([]*common.RawConfig, 0)

	allConfigs = append(allConfigs, mixed.DirectConfigs...)

	for _, subURL := range mixed.Subscriptions {
		logger.Infof("parser", "fetching nested subscription: %s", subURL)
		subText, err := fetchWithRetry(subURL)
		if err != nil {
			logger.Warnf("parser", "fetch nested sub %s: %v", subURL, err)
			continue
		}

		subText = tryDecodeWholeBody(subText)
		subURIs := Extract(subText)

		if len(subURIs) > 0 {
			subConfigs, err := ParseURIs(subURIs)
			if err != nil {
				logger.Warnf("parser", "parse nested sub %s: %v", subURL, err)
				continue
			}
			allConfigs = append(allConfigs, subConfigs...)
		}
	}

	logger.Infof("parser", "total configs extracted: %d", len(allConfigs))
	return allConfigs, nil
}
