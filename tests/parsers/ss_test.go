package parsers_test

import (
	"encoding/base64"
	"testing"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/parsers"
)

func TestSS_SIP002_Valid(t *testing.T) {
	// SIP002: ss://base64(method:password)@host:port#tag
	userinfo := base64.StdEncoding.EncodeToString([]byte("chacha20-ietf-poly1305:mysecretpassword"))
	uri := "ss://" + userinfo + "@ss.example.com:8388#MySSServer"

	p := parsers.GetParser(common.Shadowsocks)
	if p == nil {
		t.Fatal("ss parser not registered")
	}

	cfg, err := p.ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Protocol != common.Shadowsocks {
		t.Errorf("expected ss, got %s", cfg.Protocol)
	}
	if cfg.Method != "chacha20-ietf-poly1305" {
		t.Errorf("unexpected method: %s", cfg.Method)
	}
	if cfg.Password != "mysecretpassword" {
		t.Errorf("unexpected password: %s", cfg.Password)
	}
	if cfg.Server != "ss.example.com" {
		t.Errorf("unexpected server: %s", cfg.Server)
	}
	if cfg.Port != 8388 {
		t.Errorf("unexpected port: %d", cfg.Port)
	}
	if cfg.Name != "MySSServer" {
		t.Errorf("unexpected name: %s", cfg.Name)
	}
}

func TestSS_Legacy_Valid(t *testing.T) {
	// Legacy: ss://base64(method:password@host:port)#tag
	content := "aes-256-gcm:pass123@legacy.example.com:9999"
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	uri := "ss://" + encoded + "#LegacyServer"

	p := parsers.GetParser(common.Shadowsocks)
	cfg, err := p.ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Method != "aes-256-gcm" {
		t.Errorf("unexpected method: %s", cfg.Method)
	}
	if cfg.Password != "pass123" {
		t.Errorf("unexpected password: %s", cfg.Password)
	}
	if cfg.Server != "legacy.example.com" {
		t.Errorf("unexpected server: %s", cfg.Server)
	}
	if cfg.Port != 9999 {
		t.Errorf("unexpected port: %d", cfg.Port)
	}
	if cfg.Name != "LegacyServer" {
		t.Errorf("unexpected name: %s", cfg.Name)
	}
}

func TestSS_PasswordWithColon(t *testing.T) {
	// Password contains a colon — should still work (split on first : only)
	userinfo := base64.StdEncoding.EncodeToString([]byte("chacha20-ietf-poly1305:pass:with:colons"))
	uri := "ss://" + userinfo + "@server.test:8388#ColonPass"

	p := parsers.GetParser(common.Shadowsocks)
	cfg, err := p.ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Password != "pass:with:colons" {
		t.Errorf("expected password with colons, got: %s", cfg.Password)
	}
}

func TestSS_DefaultPort(t *testing.T) {
	// No port in legacy format -> default 8388
	content := "aes-128-gcm:secret@noport.example.com"
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	uri := "ss://" + encoded + "#NoPort"

	p := parsers.GetParser(common.Shadowsocks)
	cfg, err := p.ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 8388 {
		t.Errorf("expected default port 8388, got %d", cfg.Port)
	}
}
