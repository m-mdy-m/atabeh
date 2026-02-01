package parsers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

const subTag = "subscription"

func FetchSubscription(subscriptionURL string) ([]*common.RawConfig, error) {
	logger.Infof(subTag, "fetching subscription: %s", subscriptionURL)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Get(subscriptionURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscription returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Debugf(subTag, "fetched %d bytes from subscription", len(body))

	content := strings.TrimSpace(string(body))

	uris, err := TryDecodeBase64Block(content)
	if err != nil {
		logger.Debugf(subTag, "base64 decode failed, trying direct URI extraction")
		uris = ExtractURIs(content)
	}

	if len(uris) == 0 {
		logger.Warnf(subTag, "no configs found in subscription")
		return nil, nil
	}

	logger.Infof(subTag, "found %d URI(s) in subscription", len(uris))

	configs, err := ParseAll(uris)
	if err != nil {
		return nil, err
	}
	for _, cfg := range configs {
		cfg.Source = "subscription:" + subscriptionURL
	}

	return configs, nil
}
