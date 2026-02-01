package parsers_test

import (
	"testing"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/parsers"
)

func TestVLESS_ParseExampleURI(t *testing.T) {
	uri := "vless://829658bf-03c4-4c28-81e9-dd6ea141b2d0@" +
		"188.114.98.0:443" +
		"?type=ws" +
		"&security=tls" +
		"&path=%2F%3Fed%3D2560" +
		"&alpn=http%2F1.1" +
		"&encryption=none" +
		"&insecure=0" +
		"&allowInsecure=0" +
		"&host=5jq7fvwpqt5owo2fi198sa6qoxznkzfea7en4m3xroeqrt3u3q.zjde5.de5.net" +
		"&sni=5jq7fvwpqt5owo2fi198sa6qoxznkzfea7en4m3xroeqrt3u3q.zjde5.de5.net" +
		"&fp=chrome" +
		"#jabeh_farsi"

	p := parsers.GetParser(common.Vless)
	if p == nil {
		t.Fatal("no parser registered for vless")
	}

	cfg, err := p.ParseURI(uri)
	if err != nil {
		t.Fatalf("ParseURI error: %v", err)
	}

	if cfg.Name != "jabeh_farsi" {
		t.Errorf("expected name=jabeh_farsi, got %q", cfg.Name)
	}
	if cfg.Server != "188.114.98.0" {
		t.Errorf("unexpected server: %s", cfg.Server)
	}
	if cfg.Port != 443 {
		t.Errorf("unexpected port: %d", cfg.Port)
	}
	if cfg.Transport != common.WS {
		t.Errorf("expected transport=ws, got %s", cfg.Transport)
	}
	if cfg.Security != "tls" {
		t.Errorf("expected security=tls, got %s", cfg.Security)
	}
	if cfg.Extra["sni"] != "5jq7fvwpqt5owo2fi198sa6qoxznkzfea7en4m3xroeqrt3u3q.zjde5.de5.net" {
		t.Errorf("unexpected sni in extra: %s", cfg.Extra["sni"])
	}
	if cfg.Extra["fp"] != "chrome" {
		t.Errorf("unexpected fp in extra: %s", cfg.Extra["fp"])
	}
}
func TestVLESS_ValidURI(t *testing.T) {
	uri := "vless://550e8400-e29b-41d4-a716-446655440000@example.com:443?type=tcp&security=tls&sni=example.com#MyServer"

	p := parsers.GetParser(common.Vless)
	if p == nil {
		t.Fatal("vless parser not registered")
	}

	cfg, err := p.ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Protocol != common.Vless {
		t.Errorf("expected protocol vless, got %s", cfg.Protocol)
	}
	if cfg.UUID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("unexpected UUID: %s", cfg.UUID)
	}
	if cfg.Server != "example.com" {
		t.Errorf("unexpected server: %s", cfg.Server)
	}
	if cfg.Port != 443 {
		t.Errorf("unexpected port: %d", cfg.Port)
	}
	if cfg.Name != "MyServer" {
		t.Errorf("unexpected name: %s", cfg.Name)
	}
	if cfg.Transport != common.TCP {
		t.Errorf("unexpected transport: %s", cfg.Transport)
	}
	if cfg.Security != "tls" {
		t.Errorf("unexpected security: %s", cfg.Security)
	}
	if cfg.Extra["sni"] != "example.com" {
		t.Errorf("unexpected sni in extra: %s", cfg.Extra["sni"])
	}
}

func TestVLESS_RealityConfig(t *testing.T) {
	uri := "vless://uuid123@1.2.3.4:443?type=tcp&security=reality&pbkey=abc&sid=def&fp=chrome&spx=www.microsoft.com#Reality-Server"

	p := parsers.GetParser(common.Vless)
	cfg, err := p.ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Security != "reality" {
		t.Errorf("expected security=reality, got %s", cfg.Security)
	}
	if cfg.Extra["pbkey"] != "abc" {
		t.Errorf("expected pbkey=abc in extra")
	}
	if cfg.Extra["sid"] != "def" {
		t.Errorf("expected sid=def in extra")
	}
	if cfg.Extra["fp"] != "chrome" {
		t.Errorf("expected fp=chrome in extra")
	}
	if cfg.Extra["spx"] != "www.microsoft.com" {
		t.Errorf("expected spx in extra")
	}
}

func TestVLESS_DefaultPort(t *testing.T) {
	// No port specified -> should default to 443
	uri := "vless://my-uuid@server.example.com?security=tls#NoPort"

	p := parsers.GetParser(common.Vless)
	cfg, err := p.ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 443 {
		t.Errorf("expected default port 443, got %d", cfg.Port)
	}
}

func TestVLESS_MissingUUID(t *testing.T) {
	uri := "vless://@example.com:443"

	p := parsers.GetParser(common.Vless)
	_, err := p.ParseURI(uri)
	if err == nil {
		t.Fatal("expected error for missing UUID, got nil")
	}
}

func TestVLESS_MissingServer(t *testing.T) {
	uri := "vless://some-uuid@:443"

	p := parsers.GetParser(common.Vless)
	_, err := p.ParseURI(uri)
	if err == nil {
		t.Fatal("expected error for missing server, got nil")
	}
}

func TestVLESS_WebSocketTransport(t *testing.T) {
	uri := "vless://uuid@cdn.example.com:443?type=ws&security=tls&path=/secret&sni=cdn.example.com#WS-Server"

	p := parsers.GetParser(common.Vless)
	cfg, err := p.ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Transport != common.WS {
		t.Errorf("expected transport ws, got %s", cfg.Transport)
	}
	if cfg.Extra["path"] != "/secret" {
		t.Errorf("expected path=/secret in extra")
	}
}
