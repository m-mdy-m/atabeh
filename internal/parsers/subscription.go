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
	maxBodyMB    = 10
	fetchTimeout = 30 * time.Second
	maxRetries   = 3
	retryDelay   = 2 * time.Second
)

// FetchSubscription is the top-level entry for the subscription flow:
//
//	fetch → (base64 decode if needed) → Extract → ParseURIs
//
// source is stored on every returned RawConfig so storage can track origin.
func FetchSubscription(subscriptionURL string) ([]*common.RawConfig, error) {
	logger.Info("sub", "fetching "+subscriptionURL)

	body, err := fetchWithRetry(subscriptionURL)
	if err != nil {
		return nil, err
	}

	text := tryDecodeWholeBody(body)

	uris := Extract(text)
	if len(uris) == 0 {
		return nil, fmt.Errorf("no configs found in subscription")
	}
	logger.Info("sub", fmt.Sprintf("extracted %d URI(s)", len(uris)))

	configs, err := ParseURIs(uris)
	if err != nil {
		return nil, err
	}

	for _, c := range configs {
		c.Source = "subscription:" + subscriptionURL
	}
	return configs, nil
}

func fetchWithRetry(url string) (string, error) {
	client := &http.Client{
		Timeout: fetchTimeout,
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second,
		},
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			logger.Debug("sub", fmt.Sprintf("retry %d/%d", attempt, maxRetries))
			time.Sleep(retryDelay)
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("bad URL: %w", err)
		}
		req.Header.Set("User-Agent", "Atabeh/1.0")
		req.Header.Set("Accept", "*/*")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			logger.Warn("sub", fmt.Sprintf("attempt %d failed: %v", attempt, err))
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			logger.Warn("sub", fmt.Sprintf("attempt %d: HTTP %d", attempt, resp.StatusCode))
			continue
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyMB*1024*1024))
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		return string(body), nil
	}
	return "", fmt.Errorf("fetch failed after %d attempts: %w", maxRetries, lastErr)
}

func tryDecodeWholeBody(body string) string {
	trimmed := strings.TrimSpace(body)
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.RawURLEncoding,
	} {
		decoded, err := enc.DecodeString(trimmed)
		if err == nil && len(decoded) > 0 {
			return string(decoded)
		}
	}
	return body
}
