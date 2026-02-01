package parsers_test

import (
	"strings"
	"testing"

	"github.com/m-mdy-m/atabeh/internal/parsers"
)

func TestExtractURIs_SingleVLESS(t *testing.T) {
	text := "سلام! این لینک شما است:\nvless://uuid@server.com:443?security=tls#Test\nممنون"
	uris := parsers.ExtractURIs(text)

	if len(uris) != 1 {
		t.Fatalf("expected 1 URI, got %d", len(uris))
	}
	if !strings.HasPrefix(uris[0], "vless://") {
		t.Errorf("expected vless URI, got: %s", uris[0])
	}
}

func TestExtractURIs_MultipleProtocols(t *testing.T) {
	text := `🌟 کانال ما 🌟
	اینجا چند کانفیگ دارید:
	🔥 vless://uuid1@v1.example.com:443?security=tls#Server1
	⚡ vmess://eyJwcyI6InRlc3QiLCJpZCI6InVpZCIsImFkZCI6InNldnJlciIsInBvcnQiOjQ0MywibnQiOiJ0Y3AiLCJ0bHMiOiJ0bHMifQ==
	🚀 ss://Y2hhY2hhMjAtaWV0Zi1wb2x5MTMwNTpwYXNz@ss.example.com:8388#SSServer
	🎉 خوشحال شدیم!`

	uris := parsers.ExtractURIs(text)

	if len(uris) != 3 {
		t.Fatalf("expected 3 URIs, got %d: %v", len(uris), uris)
	}

	hasVless := false
	hasVmess := false
	hasSS := false
	for _, u := range uris {
		switch {
		case strings.HasPrefix(u, "vless://"):
			hasVless = true
		case strings.HasPrefix(u, "vmess://"):
			hasVmess = true
		case strings.HasPrefix(u, "ss://"):
			hasSS = true
		}
	}

	if !hasVless {
		t.Error("missing vless URI")
	}
	if !hasVmess {
		t.Error("missing vmess URI")
	}
	if !hasSS {
		t.Error("missing ss URI")
	}
}

func TestExtractURIs_EmojisAndPersian(t *testing.T) {
	text := `🔮✨🌈 سرور جدید اضافه شد! 🎯🎪🎭
	🌟⭐💫 vless://my-uuid@cdn.example.com:443?type=ws&security=tls&path=/api#فارسیسرور 🎨🎬🎤
	🏆🥇🎖️ بهترین سرور این هفته! 🎁🎀🎊`

	uris := parsers.ExtractURIs(text)

	if len(uris) != 1 {
		t.Fatalf("expected 1 URI, got %d", len(uris))
	}
	if !strings.Contains(uris[0], "cdn.example.com") {
		t.Errorf("expected server in URI: %s", uris[0])
	}
}

func TestExtractURIs_TrailingPunctuation(t *testing.T) {
	// URIs sometimes get trailing punctuation from copy-paste
	text := `لینک: vless://uuid@server.com:443?security=tls#Test.
دیگری: ss://base64data@host:8388#Name!`

	uris := parsers.ExtractURIs(text)

	for _, u := range uris {
		if strings.HasSuffix(u, ".") || strings.HasSuffix(u, "!") {
			t.Errorf("URI has trailing punctuation: %s", u)
		}
	}
}

func TestExtractURIs_Empty(t *testing.T) {
	text := "این یک متن بدون هیچ لینکی است 🎭🎪"
	uris := parsers.ExtractURIs(text)

	if len(uris) != 0 {
		t.Errorf("expected 0 URIs from plain text, got %d", len(uris))
	}
}

func TestExtractURIs_Duplicates(t *testing.T) {
	uri := "vless://uuid@server.com:443#Test"
	text := "لینک: " + uri + "\nدوباره: " + uri

	uris := parsers.ExtractURIs(text)

	if len(uris) != 1 {
		t.Errorf("expected deduplication, got %d URIs", len(uris))
	}
}
