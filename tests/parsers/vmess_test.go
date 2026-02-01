package parsers_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/parsers"
)

// helper to build vmess:// URI from a JSON map
func buildVmessURI(t *testing.T, payload map[string]any) string {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}
	return "vmess://" + base64.StdEncoding.EncodeToString(data)
}

func TestVMess_ValidConfig(t *testing.T) {
	uri := buildVmessURI(t, map[string]any{
		"ps":   "MyVMess",
		"ver":  "2",
		"id":   "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		"aid":  "0",
		"scy":  "auto",
		"add":  "vpn.example.com",
		"port": 443,
		"net":  "ws",
		"type": "none",
		"tls":  "tls",
		"path": "/ws-path",
		"host": "vpn.example.com",
	})

	p := parsers.GetParser(common.VMess)
	if p == nil {
		t.Fatal("vmess parser not registered")
	}

	cfg, err := p.ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Protocol != common.VMess {
		t.Errorf("expected vmess, got %s", cfg.Protocol)
	}
	if cfg.Name != "MyVMess" {
		t.Errorf("expected name MyVMess, got %s", cfg.Name)
	}
	if cfg.UUID != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Errorf("unexpected UUID: %s", cfg.UUID)
	}
	if cfg.Server != "vpn.example.com" {
		t.Errorf("unexpected server: %s", cfg.Server)
	}
	if cfg.Port != 443 {
		t.Errorf("unexpected port: %d", cfg.Port)
	}
	if cfg.Transport != common.WS {
		t.Errorf("expected ws transport, got %s", cfg.Transport)
	}
	if cfg.Security != "tls" {
		t.Errorf("expected tls security, got %s", cfg.Security)
	}
	if cfg.Extra["path"] != "/ws-path" {
		t.Errorf("expected path in extra")
	}
}

func TestVMess_PortAsString(t *testing.T) {
	// In the wild, port is sometimes a string
	uri := buildVmessURI(t, map[string]any{
		"ps":   "PortString",
		"id":   "test-uuid",
		"add":  "server.test",
		"port": "10443",
		"net":  "tcp",
		"tls":  "none",
	})

	p := parsers.GetParser(common.VMess)
	cfg, err := p.ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 10443 {
		t.Errorf("expected port 10443, got %d", cfg.Port)
	}
}

func TestVMess_MissingServer(t *testing.T) {
	uri := buildVmessURI(t, map[string]any{
		"ps":   "NoServer",
		"id":   "test-uuid",
		"add":  "",
		"port": 443,
	})

	p := parsers.GetParser(common.VMess)
	_, err := p.ParseURI(uri)
	if err == nil {
		t.Fatal("expected error for missing server")
	}
}

func TestVMess_H2Transport(t *testing.T) {
	uri := buildVmessURI(t, map[string]any{
		"ps":   "H2Config",
		"id":   "h2-uuid",
		"add":  "h2.example.com",
		"port": 443,
		"net":  "h2",
		"tls":  "tls",
		"path": "/h2path",
	})

	p := parsers.GetParser(common.VMess)
	cfg, err := p.ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Transport != common.H2 {
		t.Errorf("expected h2 transport, got %s", cfg.Transport)
	}
}

func TestVMess_GRPCTransport(t *testing.T) {
	uri := buildVmessURI(t, map[string]any{
		"ps":   "GRPCConfig",
		"id":   "grpc-uuid",
		"add":  "grpc.example.com",
		"port": 443,
		"net":  "grpc",
		"tls":  "tls",
	})

	p := parsers.GetParser(common.VMess)
	cfg, err := p.ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Transport != common.GRPC {
		t.Errorf("expected grpc transport, got %s", cfg.Transport)
	}
}
