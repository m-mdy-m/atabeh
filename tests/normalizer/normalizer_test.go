package normalizer_test

import (
	"testing"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/normalizer"
)

func TestNormalize_ValidVLESS(t *testing.T) {
	raw := []*common.RawConfig{
		{
			Protocol:  common.Vless,
			Name:      "TestServer",
			Server:    "vpn.example.com",
			Port:      443,
			UUID:      "test-uuid-1234",
			Transport: common.TCP,
			Security:  "tls",
		},
	}

	result, err := normalizer.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Name != "TestServer" {
		t.Errorf("expected name TestServer, got %s", result[0].Name)
	}
	if result[0].Protocol != common.Vless {
		t.Errorf("expected vless protocol")
	}
}

func TestNormalize_DuplicateRemoval(t *testing.T) {
	raw := []*common.RawConfig{
		{
			Protocol: common.Vless,
			Name:     "Server A",
			Server:   "same.server.com",
			Port:     443,
			UUID:     "same-uuid",
			Security: "tls",
		},
		{
			Protocol: common.Vless,
			Name:     "Server B (duplicate)",
			Server:   "same.server.com",
			Port:     443,
			UUID:     "same-uuid",
			Security: "tls",
		},
	}

	result, err := normalizer.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 after dedup, got %d", len(result))
	}
}

func TestNormalize_MissingServer(t *testing.T) {
	raw := []*common.RawConfig{
		{
			Protocol: common.Vless,
			Name:     "NoServer",
			Server:   "",
			Port:     443,
			UUID:     "uuid",
		},
	}

	result, _ := normalizer.Normalize(raw)
	if len(result) != 0 {
		t.Error("expected invalid config to be skipped")
	}
}

func TestNormalize_InvalidPort(t *testing.T) {
	raw := []*common.RawConfig{
		{
			Protocol: common.Vless,
			Name:     "BadPort",
			Server:   "server.com",
			Port:     0,
			UUID:     "uuid",
		},
	}

	result, _ := normalizer.Normalize(raw)
	if len(result) != 0 {
		t.Error("expected config with port 0 to be skipped")
	}
}

func TestNormalize_MissingUUID(t *testing.T) {
	raw := []*common.RawConfig{
		{
			Protocol: common.Vless,
			Name:     "NoUUID",
			Server:   "server.com",
			Port:     443,
			UUID:     "",
		},
	}

	result, _ := normalizer.Normalize(raw)
	if len(result) != 0 {
		t.Error("expected vless without UUID to be skipped")
	}
}

func TestNormalize_SSMissingPassword(t *testing.T) {
	raw := []*common.RawConfig{
		{
			Protocol: common.Shadowsocks,
			Name:     "NoPass",
			Server:   "server.com",
			Port:     8388,
			Method:   "chacha20-ietf-poly1305",
			Password: "",
		},
	}

	result, _ := normalizer.Normalize(raw)
	if len(result) != 0 {
		t.Error("expected ss without password to be skipped")
	}
}

func TestNormalize_SSMissingMethod(t *testing.T) {
	raw := []*common.RawConfig{
		{
			Protocol: common.Shadowsocks,
			Name:     "NoMethod",
			Server:   "server.com",
			Port:     8388,
			Password: "pass",
			Method:   "",
		},
	}

	result, _ := normalizer.Normalize(raw)
	if len(result) != 0 {
		t.Error("expected ss without method to be skipped")
	}
}

func TestNormalize_EmptyNameFallback(t *testing.T) {
	raw := []*common.RawConfig{
		{
			Protocol: common.Vless,
			Name:     "",
			Server:   "myserver.com",
			Port:     443,
			UUID:     "uuid-123",
			Security: "tls",
		},
	}

	result, err := normalizer.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result")
	}
	// should generate fallback name like "vless-myserver.com"
	if result[0].Name == "" {
		t.Error("expected fallback name, got empty")
	}
}

func TestNormalize_DefaultTransportAndSecurity(t *testing.T) {
	raw := []*common.RawConfig{
		{
			Protocol: common.Vless,
			Name:     "Defaults",
			Server:   "server.com",
			Port:     443,
			UUID:     "uuid",
			// Transport and Security left empty
		},
	}

	result, err := normalizer.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].Transport != common.TCP {
		t.Errorf("expected default transport tcp, got %s", result[0].Transport)
	}
	if result[0].Security != "none" {
		t.Errorf("expected default security none, got %s", result[0].Security)
	}
}

func TestNormalize_NameCleanup(t *testing.T) {
	raw := []*common.RawConfig{
		{
			Protocol: common.Vless,
			Name:     "«سرور تست»",
			Server:   "server.com",
			Port:     443,
			UUID:     "uuid",
		},
	}

	result, err := normalizer.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// guillemets should be stripped
	if result[0].Name == "«سرور تست»" {
		t.Error("expected name cleanup to strip guillemets")
	}
}
