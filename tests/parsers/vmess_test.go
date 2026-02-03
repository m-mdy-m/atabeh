package parsers_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/m-mdy-m/atabeh/internal/common"
)

func vmessURI(t *testing.T, payload map[string]any) string {
	t.Helper()
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}
	return "vmess://" + base64.StdEncoding.EncodeToString(b)
}

func basePayload() map[string]any {
	return map[string]any{
		"ps":   "test-server",
		"ver":  "2",
		"id":   "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		"aid":  "0",
		"scy":  "auto",
		"add":  "vpn.example.com",
		"port": 443,
		"net":  "tcp",
		"type": "none",
		"tls":  "tls",
	}
}

func merge(base map[string]any, override map[string]any) map[string]any {
	for k, v := range override {
		base[k] = v
	}
	return base
}

func TestVMess_BasicTCP(t *testing.T) {
	cfg, err := MustGetParser(t, common.VMess).ParseURI(vmessURI(t, basePayload()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Protocol != common.VMess {
		t.Errorf("protocol: got %q", cfg.Protocol)
	}
	if cfg.Name != "test-server" {
		t.Errorf("name: got %q", cfg.Name)
	}
	if cfg.UUID != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Errorf("uuid: got %q", cfg.UUID)
	}
	if cfg.Server != "vpn.example.com" {
		t.Errorf("server: got %q", cfg.Server)
	}
	if cfg.Port != 443 {
		t.Errorf("port: got %d", cfg.Port)
	}
	if cfg.Transport != common.TCP {
		t.Errorf("transport: got %q", cfg.Transport)
	}
	if cfg.Security != "tls" {
		t.Errorf("security: got %q", cfg.Security)
	}
}

func TestVMess_WebSocket(t *testing.T) {
	p := merge(basePayload(), map[string]any{
		"net":  "ws",
		"path": "/secret-path",
		"host": "cdn.example.com",
	})
	cfg, err := MustGetParser(t, common.VMess).ParseURI(vmessURI(t, p))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Transport != common.WS {
		t.Errorf("transport: got %q, want ws", cfg.Transport)
	}
	if cfg.Extra["path"] != "/secret-path" {
		t.Errorf("path: got %q", cfg.Extra["path"])
	}
	if cfg.Extra["host"] != "cdn.example.com" {
		t.Errorf("host: got %q", cfg.Extra["host"])
	}
}

func TestVMess_H2Transport(t *testing.T) {
	p := merge(basePayload(), map[string]any{"net": "h2", "path": "/h2"})
	cfg, err := MustGetParser(t, common.VMess).ParseURI(vmessURI(t, p))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Transport != common.H2 {
		t.Errorf("transport: got %q, want h2", cfg.Transport)
	}
}

func TestVMess_gRPCTransport(t *testing.T) {
	p := merge(basePayload(), map[string]any{"net": "grpc"})
	cfg, err := MustGetParser(t, common.VMess).ParseURI(vmessURI(t, p))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Transport != common.GRPC {
		t.Errorf("transport: got %q, want grpc", cfg.Transport)
	}
}

func TestVMess_KCPMapsToUDP(t *testing.T) {

	p := merge(basePayload(), map[string]any{"net": "kcp"})
	cfg, err := MustGetParser(t, common.VMess).ParseURI(vmessURI(t, p))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Transport != common.UDP {
		t.Errorf("transport: got %q, want udp (from kcp)", cfg.Transport)
	}
}

func TestVMess_PortAsString(t *testing.T) {
	p := merge(basePayload(), map[string]any{"port": "10443"})
	cfg, err := MustGetParser(t, common.VMess).ParseURI(vmessURI(t, p))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 10443 {
		t.Errorf("port: got %d, want 10443", cfg.Port)
	}
}

func TestVMess_PortMissing_DefaultTo443(t *testing.T) {
	p := basePayload()
	delete(p, "port")
	cfg, err := MustGetParser(t, common.VMess).ParseURI(vmessURI(t, p))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 443 {
		t.Errorf("port: got %d, want default 443", cfg.Port)
	}
}

func TestVMess_PortInvalidString(t *testing.T) {
	p := merge(basePayload(), map[string]any{"port": "not-a-number"})
	_, err := MustGetParser(t, common.VMess).ParseURI(vmessURI(t, p))
	if err == nil {
		t.Fatal("expected error for non-numeric port string")
	}
}

func TestVMess_AltID_NonZero_InExtra(t *testing.T) {
	p := merge(basePayload(), map[string]any{"aid": "64"})
	cfg, err := MustGetParser(t, common.VMess).ParseURI(vmessURI(t, p))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Extra["aid"] != "64" {
		t.Errorf("aid: got %q, want \"64\"", cfg.Extra["aid"])
	}
}

func TestVMess_AltID_Zero_NotInExtra(t *testing.T) {
	p := merge(basePayload(), map[string]any{"aid": "0"})
	cfg, err := MustGetParser(t, common.VMess).ParseURI(vmessURI(t, p))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cfg.Extra["aid"]; ok {
		t.Error("aid=0 should not appear in Extra")
	}
}

func TestVMess_CamouflageType(t *testing.T) {
	p := merge(basePayload(), map[string]any{"type": "http"})
	cfg, err := MustGetParser(t, common.VMess).ParseURI(vmessURI(t, p))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Extra["camouflage"] != "http" {
		t.Errorf("camouflage: got %q, want http", cfg.Extra["camouflage"])
	}
}

func TestVMess_EncryptionFieldPreserved(t *testing.T) {
	p := merge(basePayload(), map[string]any{"scy": "chacha20-poly1305"})
	cfg, err := MustGetParser(t, common.VMess).ParseURI(vmessURI(t, p))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Extra["encryption"] != "chacha20-poly1305" {
		t.Errorf("encryption: got %q", cfg.Extra["encryption"])
	}
}

func TestVMess_MissingServer(t *testing.T) {
	p := merge(basePayload(), map[string]any{"add": ""})
	_, err := MustGetParser(t, common.VMess).ParseURI(vmessURI(t, p))
	if err == nil {
		t.Fatal("expected error for empty server")
	}
}

func TestVMess_InvalidBase64(t *testing.T) {
	_, err := MustGetParser(t, common.VMess).ParseURI("vmess://!!!not-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestVMess_InvalidJSON(t *testing.T) {

	bad := base64.StdEncoding.EncodeToString([]byte("this is not json"))
	_, err := MustGetParser(t, common.VMess).ParseURI("vmess://" + bad)
	if err == nil {
		t.Fatal("expected error for invalid JSON payload")
	}
}

func TestVMess_URLEncodedBase64(t *testing.T) {

	p := basePayload()
	b, _ := json.Marshal(p)
	urlB64 := base64.URLEncoding.EncodeToString(b)

	cfg, err := MustGetParser(t, common.VMess).ParseURI("vmess://" + urlB64)
	if err != nil {
		t.Fatalf("unexpected error with URL-encoded base64: %v", err)
	}
	if cfg.Server != "vpn.example.com" {
		t.Errorf("server: got %q", cfg.Server)
	}
}

func TestVMess_RawBase64NoPadding(t *testing.T) {
	p := basePayload()
	b, _ := json.Marshal(p)
	rawB64 := base64.RawStdEncoding.EncodeToString(b)

	cfg, err := MustGetParser(t, common.VMess).ParseURI("vmess://" + rawB64)
	if err != nil {
		t.Fatalf("unexpected error with raw base64 (no padding): %v", err)
	}
	if cfg.Server != "vpn.example.com" {
		t.Errorf("server: got %q", cfg.Server)
	}
}
