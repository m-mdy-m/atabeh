package tester

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
)

func DialWithContext(ctx context.Context, network, addr string, cfg *common.NormalizedConfig) (net.Conn, string, error) {
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	conn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, TagConnectionError(err), err
	}

	if NeedsTLS(cfg) {
		tlsConn := tls.Client(conn, &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         GetSNI(cfg),
		})

		if err := tlsConn.HandshakeContext(ctx); err != nil {
			conn.Close()
			return nil, "tls-fail", fmt.Errorf("TLS handshake: %w", err)
		}

		return tlsConn, "", nil
	}

	return conn, "", nil
}

func NeedsTLS(cfg *common.NormalizedConfig) bool {
	sec := strings.ToLower(cfg.Security)
	return sec == "tls" || sec == "reality"
}

func GetSNI(cfg *common.NormalizedConfig) string {
	if cfg.Extra != nil {
		if sni, ok := cfg.Extra["sni"]; ok && sni != "" {
			return sni
		}
	}
	return cfg.Server
}

func TagConnectionError(err error) string {
	errStr := err.Error()

	if strings.Contains(errStr, "timeout") {
		return "timeout"
	}
	if strings.Contains(errStr, "refused") {
		return "refused"
	}
	if strings.Contains(errStr, "no route") {
		return "no-route"
	}

	if strings.Contains(errStr, "no such host") || strings.Contains(errStr, "dns") {
		return "dns-fail"
	}

	if strings.Contains(errStr, "reset") || strings.Contains(errStr, "broken pipe") {
		return "dpi-reset"
	}

	return "network-fail"
}
