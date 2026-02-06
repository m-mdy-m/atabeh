package tester

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

func validateRealConnection(addr string, cfg *common.NormalizedConfig) bool {

	if !testBandwidth(addr, cfg) {
		logger.Debugf("tester", "[%s] failed bandwidth test", cfg.Name)
		return false
	}

	if !testPersistence(addr, cfg) {
		logger.Debugf("tester", "[%s] failed persistence test", cfg.Name)
		return false
	}

	return true
}

func testBandwidth(addr string, cfg *common.NormalizedConfig) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testURL := "http://www.gstatic.com/generate_204"

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				dialer := &net.Dialer{
					Timeout: 5 * time.Second,
				}
				return dialer.DialContext(ctx, network, addr)
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return false
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	bytesRead, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		return false
	}

	duration := time.Since(start)

	if duration.Seconds() == 0 {
		return false
	}

	kbps := float64(bytesRead) / 1024.0 / duration.Seconds()

	minBandwidth := 50.0
	return kbps >= minBandwidth
}

func testPersistence(addr string, cfg *common.NormalizedConfig) bool {
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return false
	}
	defer conn.Close()

	time.Sleep(30 * time.Second)

	conn.SetDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Write([]byte("test"))
	if err != nil {
		return false
	}

	return true
}
