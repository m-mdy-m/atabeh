package parsers_test

import (
	"testing"

	"github.com/m-mdy-m/atabeh/internal/common"
)

func TestTrojan_BasicTLS(t *testing.T) {
	uri := "trojan://mypassword@trojan.example.com:443#BasicTrojan"

	cfg, err := MustGetParser(t, common.Trojan).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Protocol != common.Trojan {
		t.Errorf("protocol: got %q", cfg.Protocol)
	}
	if cfg.Password != "mypassword" {
		t.Errorf("password: got %q", cfg.Password)
	}
	if cfg.Server != "trojan.example.com" {
		t.Errorf("server: got %q", cfg.Server)
	}
	if cfg.Port != 443 {
		t.Errorf("port: got %d", cfg.Port)
	}
	if cfg.Name != "BasicTrojan" {
		t.Errorf("name: got %q", cfg.Name)
	}

	if cfg.Security != "tls" {
		t.Errorf("security: got %q, want tls (default)", cfg.Security)
	}
	if cfg.Transport != common.TCP {
		t.Errorf("transport: got %q, want tcp (default)", cfg.Transport)
	}
}

func TestTrojan_WebSocket(t *testing.T) {
	uri := "trojan://pw123@ws.trojan.com:443?type=ws&security=tls&path=/ws#WSTrojan"

	cfg, err := MustGetParser(t, common.Trojan).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Transport != common.WS {
		t.Errorf("transport: got %q, want ws", cfg.Transport)
	}
	if cfg.Extra["path"] != "/ws" {
		t.Errorf("path: got %q", cfg.Extra["path"])
	}
}

func TestTrojan_gRPC(t *testing.T) {
	uri := "trojan://pass@grpc.trojan.com:443?type=grpc&security=tls&serviceName=svc#gRPC"

	cfg, err := MustGetParser(t, common.Trojan).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Transport != common.GRPC {
		t.Errorf("transport: got %q, want grpc", cfg.Transport)
	}
	if cfg.Extra["serviceName"] != "svc" {
		t.Errorf("serviceName: got %q", cfg.Extra["serviceName"])
	}
}

func TestTrojan_NonDefaultPort(t *testing.T) {
	uri := "trojan://pass@trojan.test:8443#CustomPort"

	cfg, err := MustGetParser(t, common.Trojan).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8443 {
		t.Errorf("port: got %d, want 8443", cfg.Port)
	}
}

func TestTrojan_DefaultPort443(t *testing.T) {

	uri := "trojan://pass@trojan.test#NoPort"

	cfg, err := MustGetParser(t, common.Trojan).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 443 {
		t.Errorf("port: got %d, want default 443", cfg.Port)
	}
}

func TestTrojan_URLEncodedPassword(t *testing.T) {

	uri := "trojan://p%40ss%3Aword%21@trojan.test:443#Encoded"

	cfg, err := MustGetParser(t, common.Trojan).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Password != "p@ss:word!" {
		t.Errorf("password: got %q, want p@ss:word!", cfg.Password)
	}
}

func TestTrojan_ExtraParamsPreserved(t *testing.T) {
	uri := "trojan://pass@trojan.test:443?type=tcp&security=tls&sni=trojan.test&alpn=h2,http/1.1#Extras"

	cfg, err := MustGetParser(t, common.Trojan).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cfg.Extra["type"]; ok {
		t.Error("'type' should not appear in Extra")
	}
	if _, ok := cfg.Extra["security"]; ok {
		t.Error("'security' should not appear in Extra")
	}

	if cfg.Extra["sni"] != "trojan.test" {
		t.Errorf("sni: got %q", cfg.Extra["sni"])
	}
	if cfg.Extra["alpn"] != "h2,http/1.1" {
		t.Errorf("alpn: got %q", cfg.Extra["alpn"])
	}
}

func TestTrojan_MissingPassword(t *testing.T) {
	uri := "trojan://@trojan.test:443#NoPass"
	_, err := MustGetParser(t, common.Trojan).ParseURI(uri)
	if err == nil {
		t.Fatal("expected error for missing password")
	}
}

func TestTrojan_MissingHost(t *testing.T) {
	uri := "trojan://pass@:443#NoHost"
	_, err := MustGetParser(t, common.Trojan).ParseURI(uri)
	if err == nil {
		t.Fatal("expected error for missing host")
	}
}

func TestTrojan_WrongScheme(t *testing.T) {
	uri := "vless://pass@trojan.test:443"
	_, err := MustGetParser(t, common.Trojan).ParseURI(uri)
	if err == nil {
		t.Fatal("expected error for wrong scheme")
	}
}

func TestTrojan_NonNumericPort(t *testing.T) {
	uri := "trojan://pass@trojan.test:abc#BadPort"
	_, err := MustGetParser(t, common.Trojan).ParseURI(uri)
	if err == nil {
		t.Fatal("expected error for non-numeric port")
	}
}
