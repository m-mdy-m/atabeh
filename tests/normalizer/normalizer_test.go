package normalizer_test

import (
	"testing"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/normalizer"
)

func validVless(overrides ...func(*common.RawConfig)) *common.RawConfig {
	c := &common.RawConfig{
		Protocol:  common.Vless,
		Name:      "TestServer",
		Server:    "vpn.example.com",
		Port:      443,
		UUID:      "550e8400-e29b-41d4-a716-446655440000",
		Transport: common.TCP,
		Security:  "tls",
	}
	for _, fn := range overrides {
		fn(c)
	}
	return c
}

func validSS() *common.RawConfig {
	return &common.RawConfig{
		Protocol: common.Shadowsocks,
		Name:     "SS-Server",
		Server:   "ss.example.com",
		Port:     8388,
		Password: "s3cret",
		Method:   "chacha20-ietf-poly1305",
	}
}

func TestNormalize_ValidVless_PassesThrough(t *testing.T) {
	out, err := normalizer.Normalize([]*common.RawConfig{validVless()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if out[0].Name != "TestServer" {
		t.Errorf("name: got %q", out[0].Name)
	}
	if out[0].Protocol != common.Vless {
		t.Errorf("protocol: got %q", out[0].Protocol)
	}
	if out[0].Server != "vpn.example.com" {
		t.Errorf("server: got %q", out[0].Server)
	}
}

func TestNormalize_ValidSS_PassesThrough(t *testing.T) {
	out, err := normalizer.Normalize([]*common.RawConfig{validSS()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1, got %d", len(out))
	}
	if out[0].Method != "chacha20-ietf-poly1305" {
		t.Errorf("method: got %q", out[0].Method)
	}
}

func TestNormalize_DuplicatesRemoved(t *testing.T) {
	a := validVless(func(c *common.RawConfig) { c.Name = "First" })
	b := validVless(func(c *common.RawConfig) { c.Name = "Second (dup)" })

	out, err := normalizer.Normalize([]*common.RawConfig{a, b})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 {
		t.Errorf("expected 1 after dedup, got %d", len(out))
	}
	if out[0].Name != "First" {
		t.Errorf("should keep first occurrence, got name=%q", out[0].Name)
	}
}

func TestNormalize_DifferentUUID_NotDup(t *testing.T) {
	a := validVless(func(c *common.RawConfig) { c.UUID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" })
	b := validVless(func(c *common.RawConfig) { c.UUID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" })

	out, _ := normalizer.Normalize([]*common.RawConfig{a, b})
	if len(out) != 2 {
		t.Errorf("different UUIDs → not duplicates, expected 2, got %d", len(out))
	}
}

func TestNormalize_DifferentPort_NotDup(t *testing.T) {
	a := validVless(func(c *common.RawConfig) { c.Port = 443 })
	b := validVless(func(c *common.RawConfig) { c.Port = 8443 })

	out, _ := normalizer.Normalize([]*common.RawConfig{a, b})
	if len(out) != 2 {
		t.Errorf("different ports → not duplicates, expected 2, got %d", len(out))
	}
}

func TestNormalize_EmptyServer_Skipped(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Server = "" })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 0 {
		t.Error("empty server should be rejected")
	}
}

func TestNormalize_PrivateIP_Skipped(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Server = "192.168.1.1" })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 0 {
		t.Error("private IP should be rejected by default")
	}
}

func TestNormalize_LoopbackIP_Skipped(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Server = "127.0.0.1" })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 0 {
		t.Error("loopback IP should be rejected")
	}
}

func TestNormalize_PublicIP_Accepted(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Server = "8.8.8.8" })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 1 {
		t.Error("public IP should be accepted")
	}
}

func TestNormalize_Port0_Skipped(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Port = 0 })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 0 {
		t.Error("port 0 should be rejected")
	}
}

func TestNormalize_Port_Negative_Skipped(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Port = -1 })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 0 {
		t.Error("negative port should be rejected")
	}
}

func TestNormalize_Port_65536_Skipped(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Port = 65536 })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 0 {
		t.Error("port > 65535 should be rejected")
	}
}

func TestNormalize_Port_65535_Accepted(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Port = 65535 })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 1 {
		t.Error("port 65535 should be accepted")
	}
}

func TestNormalize_Port1_Accepted(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Port = 1 })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 1 {
		t.Error("port 1 should be accepted")
	}
}

func TestNormalize_MissingUUID_Skipped(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.UUID = "" })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 0 {
		t.Error("vless without UUID should be rejected")
	}
}

func TestNormalize_InvalidUUIDFormat_Skipped(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.UUID = "not-a-valid-uuid" })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 0 {
		t.Error("invalid UUID format should be rejected")
	}
}

func TestNormalize_ValidUUID_Accepted(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.UUID = "12345678-1234-1234-1234-123456789abc" })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 1 {
		t.Error("valid UUID should be accepted")
	}
}

func TestNormalize_SS_MissingPassword_Skipped(t *testing.T) {
	c := validSS()
	c.Password = ""
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 0 {
		t.Error("SS without password should be rejected")
	}
}

func TestNormalize_SS_MissingMethod_Skipped(t *testing.T) {
	c := validSS()
	c.Method = ""
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 0 {
		t.Error("SS without method should be rejected")
	}
}

func TestNormalize_SS_UnsupportedMethod_Skipped(t *testing.T) {
	c := validSS()
	c.Method = "rc4"
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 0 {
		t.Error("SS with unsupported method should be rejected")
	}
}

func TestNormalize_SS_AllValidMethods_Accepted(t *testing.T) {
	methods := []string{
		"aes-128-gcm",
		"aes-256-gcm",
		"chacha20-ietf-poly1305",
		"xchacha20-ietf-poly1305",
		"2022-blake3-aes-128-gcm",
		"2022-blake3-aes-256-gcm",
	}
	for _, m := range methods {
		c := validSS()
		c.Method = m
		out, _ := normalizer.Normalize([]*common.RawConfig{c})
		if len(out) != 1 {
			t.Errorf("method %q should be accepted", m)
		}
	}
}

func TestNormalize_EmptyTransport_DefaultsTCP(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Transport = "" })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) == 0 {
		t.Fatal("expected config to pass validation")
	}
	if out[0].Transport != common.TCP {
		t.Errorf("default transport: got %q, want tcp", out[0].Transport)
	}
}

func TestNormalize_SS_EmptyTransport_DefaultsUDP(t *testing.T) {
	c := validSS()
	c.Transport = ""
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) == 0 {
		t.Fatal("expected config to pass validation")
	}
	if out[0].Transport != common.UDP {
		t.Errorf("SS default transport: got %q, want udp", out[0].Transport)
	}
}

func TestNormalize_EmptySecurity_DefaultsNone(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Security = "" })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) == 0 {
		t.Fatal("expected config to pass validation")
	}
	if out[0].Security != "none" {
		t.Errorf("default security: got %q, want none", out[0].Security)
	}
}

func TestNormalize_GuillemetsStripped(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Name = "«سرور تست»" })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) == 0 {
		t.Fatal("config should pass")
	}
	if out[0].Name == "«سرور تست»" {
		t.Error("guillemets should be stripped from name")
	}
}

func TestNormalize_EmojiStripped(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Name = "🔥Server🔥" })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) == 0 {
		t.Fatal("config should pass")
	}
	if out[0].Name == "🔥Server🔥" {
		t.Error("emojis should be stripped from name")
	}
}

func TestNormalize_EmptyName_FallbackGenerated(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Name = "" })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) == 0 {
		t.Fatal("config should pass")
	}
	if out[0].Name == "" {
		t.Error("empty name should get a generated fallback")
	}
}

func TestNormalize_WhitespaceOnlyName_FallbackGenerated(t *testing.T) {
	c := validVless(func(r *common.RawConfig) { r.Name = "   \t  " })
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) == 0 {
		t.Fatal("config should pass")
	}
	if out[0].Name == "" || out[0].Name == "   \t  " {
		t.Error("whitespace-only name should be replaced with fallback")
	}
}

func TestNormalize_MixedBatch(t *testing.T) {
	batch := []*common.RawConfig{
		validVless(),
		validVless(func(r *common.RawConfig) { r.Server = "" }),
		validSS(),
		validVless(func(r *common.RawConfig) { r.UUID = "bad" }),
		validVless(func(r *common.RawConfig) { r.Server = "192.168.0.1" }),
	}

	out, _ := normalizer.Normalize(batch)
	if len(out) != 2 {
		t.Errorf("expected 2 valid configs out of 5, got %d", len(out))
	}
}

func TestNormalize_UnsupportedProtocol_Skipped(t *testing.T) {
	c := &common.RawConfig{
		Protocol: "wireguard",
		Name:     "WG",
		Server:   "wg.example.com",
		Port:     51820,
	}
	out, _ := normalizer.Normalize([]*common.RawConfig{c})
	if len(out) != 0 {
		t.Error("unsupported protocol should be rejected")
	}
}
