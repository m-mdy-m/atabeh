package parsers_test

import (
	"encoding/base64"
	"testing"

	"github.com/m-mdy-m/atabeh/internal/common"
)

func TestSS_SIP002_chacha20(t *testing.T) {
	userinfo := base64.StdEncoding.EncodeToString([]byte("chacha20-ietf-poly1305:s3cret"))
	uri := "ss://" + userinfo + "@ss.example.com:8388#Chacha"

	cfg, err := MustGetParser(t, common.Shadowsocks).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Protocol != common.Shadowsocks {
		t.Errorf("protocol: got %q", cfg.Protocol)
	}
	if cfg.Method != "chacha20-ietf-poly1305" {
		t.Errorf("method: got %q", cfg.Method)
	}
	if cfg.Password != "s3cret" {
		t.Errorf("password: got %q", cfg.Password)
	}
	if cfg.Server != "ss.example.com" {
		t.Errorf("server: got %q", cfg.Server)
	}
	if cfg.Port != 8388 {
		t.Errorf("port: got %d", cfg.Port)
	}
	if cfg.Name != "Chacha" {
		t.Errorf("name: got %q", cfg.Name)
	}
	if cfg.Transport != common.UDP {
		t.Errorf("transport: got %q, want udp", cfg.Transport)
	}
}

func TestSS_SIP002_aes256gcm(t *testing.T) {
	userinfo := base64.StdEncoding.EncodeToString([]byte("aes-256-gcm:MyP@ssw0rd!"))
	uri := "ss://" + userinfo + "@10.0.0.1:9999#AES256"

	cfg, err := MustGetParser(t, common.Shadowsocks).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Method != "aes-256-gcm" {
		t.Errorf("method: got %q", cfg.Method)
	}
	if cfg.Password != "MyP@ssw0rd!" {
		t.Errorf("password: got %q", cfg.Password)
	}
	if cfg.Port != 9999 {
		t.Errorf("port: got %d", cfg.Port)
	}
}

func TestSS_SIP002_2022Blake3(t *testing.T) {

	userinfo := base64.StdEncoding.EncodeToString([]byte("2022-blake3-aes-256-gcm:longkey1234567890"))
	uri := "ss://" + userinfo + "@blake.example.com:8388#Blake3"

	cfg, err := MustGetParser(t, common.Shadowsocks).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Method != "2022-blake3-aes-256-gcm" {
		t.Errorf("method: got %q", cfg.Method)
	}
}

func TestSS_Legacy_Basic(t *testing.T) {
	content := "aes-128-gcm:pass123@legacy.example.com:9999"
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	uri := "ss://" + encoded + "#LegacyBasic"

	cfg, err := MustGetParser(t, common.Shadowsocks).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Method != "aes-128-gcm" {
		t.Errorf("method: got %q", cfg.Method)
	}
	if cfg.Password != "pass123" {
		t.Errorf("password: got %q", cfg.Password)
	}
	if cfg.Server != "legacy.example.com" {
		t.Errorf("server: got %q", cfg.Server)
	}
	if cfg.Port != 9999 {
		t.Errorf("port: got %d", cfg.Port)
	}
	if cfg.Name != "LegacyBasic" {
		t.Errorf("name: got %q", cfg.Name)
	}
}

func TestSS_Legacy_NoPort_Default8388(t *testing.T) {
	content := "aes-256-gcm:secret@noport.example.com"
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	uri := "ss://" + encoded + "#NoPort"

	cfg, err := MustGetParser(t, common.Shadowsocks).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8388 {
		t.Errorf("port: got %d, want default 8388", cfg.Port)
	}
}

func TestSS_SIP002_PasswordWithColons(t *testing.T) {

	userinfo := base64.StdEncoding.EncodeToString([]byte("chacha20-ietf-poly1305:pass:with:colons"))
	uri := "ss://" + userinfo + "@server.test:8388#ColonPass"

	cfg, err := MustGetParser(t, common.Shadowsocks).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Password != "pass:with:colons" {
		t.Errorf("password: got %q, want pass:with:colons", cfg.Password)
	}
}

func TestSS_Legacy_PasswordWithColons(t *testing.T) {
	content := "aes-256-gcm:user:pass:extra@legacy.test:1234"
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	uri := "ss://" + encoded + "#LegacyColon"

	cfg, err := MustGetParser(t, common.Shadowsocks).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Password != "user:pass:extra" {
		t.Errorf("password: got %q, want user:pass:extra", cfg.Password)
	}
	if cfg.Server != "legacy.test" {
		t.Errorf("server: got %q", cfg.Server)
	}
}

func TestSS_Legacy_IPv6Host(t *testing.T) {

	content := "aes-128-gcm:secret@[::1]:8388"
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	uri := "ss://" + encoded + "#IPv6"

	cfg, err := MustGetParser(t, common.Shadowsocks).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server != "::1" {
		t.Errorf("server: got %q, want ::1", cfg.Server)
	}
	if cfg.Port != 8388 {
		t.Errorf("port: got %d", cfg.Port)
	}
}

func TestSS_Legacy_URLEncodedName(t *testing.T) {
	content := "aes-256-gcm:pw@host.test:8388"
	encoded := base64.StdEncoding.EncodeToString([]byte(content))

	uri := "ss://" + encoded + "#%D8%B3%D8%B1%D8%B0%D8%B3%D8%AA"

	cfg, err := MustGetParser(t, common.Shadowsocks).ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name == "" {
		t.Error("name should not be empty after URL-decode")
	}
}

func TestSS_InvalidBase64(t *testing.T) {
	_, err := MustGetParser(t, common.Shadowsocks).ParseURI("ss://!!!garbage!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestSS_MissingMethodSeparator(t *testing.T) {

	userinfo := base64.StdEncoding.EncodeToString([]byte("nomethodseparator"))
	uri := "ss://" + userinfo + "@host.test:8388#Bad"

	_, err := MustGetParser(t, common.Shadowsocks).ParseURI(uri)
	if err == nil {
		t.Fatal("expected error when method:password separator is missing")
	}
}

func TestSS_Legacy_MissingAt(t *testing.T) {

	content := "aes-256-gcm:passNOHOST"
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	uri := "ss://" + encoded + "#NoAt"

	_, err := MustGetParser(t, common.Shadowsocks).ParseURI(uri)
	if err == nil {
		t.Fatal("expected error when @ is missing in legacy content")
	}
}
