package parsers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

const (
	subTag            = "subscription"
	maxSubscriptionMB = 10
	fetchTimeout      = 30 * time.Second
	maxRetries        = 3
	retryDelay        = 2 * time.Second
)

type SubscriptionFetcher struct {
	client    *http.Client
	userAgent string
}

func NewSubscriptionFetcher() *SubscriptionFetcher {
	return &SubscriptionFetcher{
		client: &http.Client{
			Timeout: fetchTimeout,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: false,
				DisableKeepAlives:  false,
			},
		},
		userAgent: "Atabeh/1.0 (VPN Config Tester)",
	}
}

func (sf *SubscriptionFetcher) FetchSubscription(subscriptionURL string) ([]*common.RawConfig, error) {
	logger.Infof(subTag, "fetching subscription: %s", subscriptionURL)

	content, err := sf.fetchWithRetry(subscriptionURL)
	if err != nil {
		return nil, err
	}

	logger.Debugf(subTag, "fetched %d bytes from subscription", len(content))

	uris, err := sf.tryDecodeBase64(content)
	if err != nil {
		logger.Debugf(subTag, "base64 decode failed, trying direct URI extraction: %v", err)

		uris = ExtractURIs(content)
	}

	if len(uris) == 0 {
		logger.Warnf(subTag, "no configs found in subscription")
		return nil, fmt.Errorf("no valid configs found in subscription")
	}

	logger.Infof(subTag, "extracted %d URI(s) from subscription", len(uris))

	configs, err := ParseAll(uris)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configs: %w", err)
	}

	for _, cfg := range configs {
		cfg.Source = "subscription:" + subscriptionURL
	}

	logger.Infof(subTag, "successfully parsed %d configs from subscription", len(configs))
	return configs, nil
}

func (sf *SubscriptionFetcher) fetchWithRetry(url string) (string, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			logger.Debugf(subTag, "retry attempt %d/%d after %v", attempt, maxRetries, retryDelay)
			time.Sleep(retryDelay)
		}

		content, err := sf.fetchOnce(url)
		if err == nil {
			return content, nil
		}

		lastErr = err
		logger.Warnf(subTag, "fetch attempt %d failed: %v", attempt, err)
	}

	return "", fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

func (sf *SubscriptionFetcher) fetchOnce(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", sf.userAgent)
	req.Header.Set("Accept", "*/*")

	resp, err := sf.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("subscription returned HTTP %d", resp.StatusCode)
	}

	limitedReader := io.LimitReader(resp.Body, maxSubscriptionMB*1024*1024)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

func (sf *SubscriptionFetcher) tryDecodeBase64(content string) ([]string, error) {
	content = strings.TrimSpace(content)

	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {

		decoded, err = base64.URLEncoding.DecodeString(content)
		if err != nil {

			decoded, err = base64.RawStdEncoding.DecodeString(content)
			if err != nil {

				decoded, err = base64.RawURLEncoding.DecodeString(content)
				if err != nil {
					return nil, fmt.Errorf("all base64 decode attempts failed: %w", err)
				}
			}
		}
	}

	decodedStr := string(decoded)
	logger.Debugf(subTag, "successfully decoded base64, length: %d", len(decodedStr))

	lines := strings.Split(strings.TrimSpace(decodedStr), "\n")
	var uris []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if isLikelyURI(line) {
			uris = append(uris, line)
		}
	}

	if len(uris) == 0 {
		uris = ExtractURIs(decodedStr)
	}

	logger.Infof(subTag, "extracted %d URI(s) from decoded base64", len(uris))
	return uris, nil
}

func isLikelyURI(s string) bool {
	prefixes := []string{"vless://", "vmess://", "ss://", "trojan://", "socks://"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func FetchSubscription(subscriptionURL string) ([]*common.RawConfig, error) {
	fetcher := NewSubscriptionFetcher()
	return fetcher.FetchSubscription(subscriptionURL)
}
