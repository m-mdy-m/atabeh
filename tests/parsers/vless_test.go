package parsers_test

import (
	"testing"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/parsers"
)

func MustGetParser(t *testing.T, kind common.Kind) parsers.Parser {
	t.Helper()
	p := parsers.GetParser(kind)
	if p == nil {
		t.Fatalf("parser for %q not registered", kind)
	}
	return p
}
func TestVLESS_BasicTLS(t *testing.T) {
	uri := "vless://550e8400-e29b-41d4-a716-446655440000@vpn.example.com:443" +
		"?type=tcp&security=tls&sni=vpn.example.com#BasicTLS"

	cfg, err := MustGetParser(t, common.Vless).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := map[string][2]string{
		"protocol":  {string(cfg.Protocol), "vless"},
		"name":      {cfg.Name, "BasicTLS"},
		"server":    {cfg.Server, "vpn.example.com"},
		"uuid":      {cfg.UUID, "550e8400-e29b-41d4-a716-446655440000"},
		"transport": {string(cfg.Transport), "tcp"},
		"security":  {cfg.Security, "tls"},
		"sni":       {cfg.Extra["sni"], "vpn.example.com"},
	}
	for field, want := range checks {
		if want[0] != want[1] {
			t.Errorf("%s: got %q, want %q", field, want[0], want[1])
		}
	}
	if cfg.Port != 443 {
		t.Errorf("port: got %d, want 443", cfg.Port)
	}
}

func TestVLESS_WebSocket_CDN(t *testing.T) {
	// Common pattern: WS over TLS behind a CDN (Cloudflare-style)
	uri := "vless://aaa-bbb-ccc@cdn.example.com:443" +
		"?type=ws&security=tls&path=%2Fapi%2Fstream&sni=cdn.example.com&fp=chrome#WS-CDN"

	cfg, err := MustGetParser(t, common.Vless).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Transport != common.WS {
		t.Errorf("transport: got %q, want ws", cfg.Transport)
	}
	if cfg.Extra["path"] != "/api/stream" {
		t.Errorf("path: got %q, want /api/stream", cfg.Extra["path"])
	}
	if cfg.Extra["fp"] != "chrome" {
		t.Errorf("fp: got %q, want chrome", cfg.Extra["fp"])
	}
}

func TestVLESS_Reality(t *testing.T) {
	// Reality configs are very common in Iran due to deep-packet-inspection bypass
	uri := "vless://real-uuid@1.2.3.4:443" +
		"?type=tcp&security=reality&pbkey=abcdef1234567890&sid=aabb&fp=chrome" +
		"&spx=www.microsoft.com#Reality-Server"

	cfg, err := MustGetParser(t, common.Vless).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Security != "reality" {
		t.Errorf("security: got %q, want reality", cfg.Security)
	}
	if cfg.Extra["pbkey"] != "abcdef1234567890" {
		t.Errorf("pbkey missing or wrong: %q", cfg.Extra["pbkey"])
	}
	if cfg.Extra["spx"] != "www.microsoft.com" {
		t.Errorf("spx: got %q", cfg.Extra["spx"])
	}
}

func TestVLESS_gRPC(t *testing.T) {
	uri := "vless://grpc-uuid@grpc.example.com:443?type=grpc&security=tls&serviceName=myService#gRPC-Server"

	cfg, err := MustGetParser(t, common.Vless).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Transport != common.GRPC {
		t.Errorf("transport: got %q, want grpc", cfg.Transport)
	}
	if cfg.Extra["serviceName"] != "myService" {
		t.Errorf("serviceName: got %q", cfg.Extra["serviceName"])
	}
}

func TestVLESS_H2(t *testing.T) {
	uri := "vless://h2-uuid@h2.example.com:443?type=h2&security=tls#H2-Server"

	cfg, err := MustGetParser(t, common.Vless).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Transport != common.H2 {
		t.Errorf("transport: got %q, want h2", cfg.Transport)
	}
}

func TestVLESS_FullRealWorld(t *testing.T) {
	// Exact-style URI pulled from a typical Iranian subscription
	uri := "vless://829658bf-03c4-4c28-81e9-dd6ea141b2d0@188.114.98.0:443" +
		"?type=ws&security=tls" +
		"&path=%2F%3Fed%3D2560" +
		"&alpn=http%2F1.1" +
		"&encryption=none" +
		"&insecure=0" +
		"&allowInsecure=0" +
		"&host=long-cdn-subdomain.example.net" +
		"&sni=long-cdn-subdomain.example.net" +
		"&fp=chrome" +
		"#jabeh_farsi"

	cfg, err := MustGetParser(t, common.Vless).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Name != "jabeh_farsi" {
		t.Errorf("name: got %q", cfg.Name)
	}
	if cfg.Server != "188.114.98.0" {
		t.Errorf("server: got %q", cfg.Server)
	}
	if cfg.Transport != common.WS {
		t.Errorf("transport: got %q", cfg.Transport)
	}
	if cfg.Extra["sni"] != "long-cdn-subdomain.example.net" {
		t.Errorf("sni: got %q", cfg.Extra["sni"])
	}
}

// ---------------------------------------------------------------------------
// Default-value behaviour
// ---------------------------------------------------------------------------

func TestVLESS_DefaultPort(t *testing.T) {
	uri := "vless://my-uuid@server.example.com?security=tls#NoPort"

	cfg, err := MustGetParser(t, common.Vless).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 443 {
		t.Errorf("port: got %d, want default 443", cfg.Port)
	}
}

func TestVLESS_DefaultTransportAndSecurity(t *testing.T) {
	// No type= or security= — should default to tcp / none
	uri := "vless://uuid-here@server.example.com:8443#Defaults"

	cfg, err := MustGetParser(t, common.Vless).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Transport != common.TCP {
		t.Errorf("transport: got %q, want tcp", cfg.Transport)
	}
	if cfg.Security != "none" {
		t.Errorf("security: got %q, want none", cfg.Security)
	}
}

// ---------------------------------------------------------------------------
// Error cases — parser must reject these cleanly
// ---------------------------------------------------------------------------

func TestVLESS_MissingUUID(t *testing.T) {
	uri := "vless://@example.com:443"
	_, err := MustGetParser(t, common.Vless).ParseURI(uri)
	if err == nil {
		t.Fatal("expected error for missing UUID")
	}
}

func TestVLESS_EmptyUUID(t *testing.T) {
	uri := "vless://:443"
	_, err := MustGetParser(t, common.Vless).ParseURI(uri)
	if err == nil {
		t.Fatal("expected error for empty UUID")
	}
}

func TestVLESS_MissingServer(t *testing.T) {
	uri := "vless://some-uuid@:443"
	_, err := MustGetParser(t, common.Vless).ParseURI(uri)
	if err == nil {
		t.Fatal("expected error for missing server")
	}
}

func TestVLESS_WrongScheme(t *testing.T) {
	uri := "vmess://550e8400-e29b-41d4-a716-446655440000@example.com:443"
	_, err := MustGetParser(t, common.Vless).ParseURI(uri)
	if err == nil {
		t.Fatal("expected error for wrong scheme")
	}
}

func TestVLESS_NonNumericPort(t *testing.T) {
	uri := "vless://uuid@example.com:abc?security=tls#Bad"
	_, err := MustGetParser(t, common.Vless).ParseURI(uri)
	if err == nil {
		t.Fatal("expected error for non-numeric port")
	}
}

// ---------------------------------------------------------------------------
// Extra fields are preserved
// ---------------------------------------------------------------------------

func TestVLESS_ExtraFieldsPreserved(t *testing.T) {
	uri := "vless://uuid@srv.com:443?type=tcp&security=tls&custom1=hello&custom2=world#Extras"

	cfg, err := MustGetParser(t, common.Vless).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// type & security are consumed — must NOT appear in Extra
	if _, ok := cfg.Extra["type"]; ok {
		t.Error("'type' should not be in Extra")
	}
	if _, ok := cfg.Extra["security"]; ok {
		t.Error("'security' should not be in Extra")
	}
	if cfg.Extra["custom1"] != "hello" {
		t.Errorf("custom1: got %q", cfg.Extra["custom1"])
	}
	if cfg.Extra["custom2"] != "world" {
		t.Errorf("custom2: got %q", cfg.Extra["custom2"])
	}
}
