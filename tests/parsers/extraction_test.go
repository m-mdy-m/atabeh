package parsers_test

import (
	"strings"
	"testing"

	"github.com/m-mdy-m/atabeh/internal/parsers"
)

func TestExtract_SingleVLESS_InPersianText(t *testing.T) {
	text := "سلام! این لینک شما است:\nvless://uuid@server.com:443?security=tls#Test\nممنون"
	uris := parsers.ExtractURIs(text)

	if len(uris) != 1 {
		t.Fatalf("expected 1 URI, got %d", len(uris))
	}
	if !strings.HasPrefix(uris[0], "vless://") {
		t.Errorf("expected vless prefix, got: %s", uris[0])
	}
}

func TestExtract_MultipleProtocols(t *testing.T) {
	text := `vless://uuid1@v1.example.com:443?security=tls#Server1
vmess://eyJwcyI6InRlc3QiLCJpZCI6InVpZCIsImFkZCI6InNldnJlciIsInBvcnQiOjQ0MywibnQiOiJ0Y3AiLCJ0bHMiOiJ0bHMifQ==
ss://Y2hhY2hhMjAtaWV0Zi1wb2x5MTMwNTpwYXNz@ss.example.com:8388#SSServer
trojan://mypass@trojan.test:443#Trojan1
socks5://proxy.test:1080#Socks1`

	uris := parsers.ExtractURIs(text)

	protocols := map[string]bool{}
	for _, u := range uris {
		for _, prefix := range []string{"vless://", "vmess://", "ss://", "trojan://", "socks5://"} {
			if strings.HasPrefix(u, prefix) {
				protocols[prefix] = true
			}
		}
	}
	for _, want := range []string{"vless://", "vmess://", "ss://", "trojan://", "socks5://"} {
		if !protocols[want] {
			t.Errorf("missing protocol %q in extracted URIs", want)
		}
	}
}

func TestExtract_EmojiHeavyTelegramPost(t *testing.T) {
	text := `🔮✨🌈 سرور جدید اضافه شد! 🎯🎪🎭
🌟⭐💫 vless://my-uuid@cdn.example.com:443?type=ws&security=tls&path=/api#فارسیسرور 🎨🎬🎤
🏆🥇🎖️ بهترین سرور این هفته! 🎁🎀🎊`

	uris := parsers.ExtractURIs(text)
	if len(uris) != 1 {
		t.Fatalf("expected 1 URI, got %d: %v", len(uris), uris)
	}
	if !strings.Contains(uris[0], "cdn.example.com") {
		t.Errorf("server not found in URI: %s", uris[0])
	}
}

func TestExtract_TelegramHelper(t *testing.T) {
	msg := "🚀 New config! 🚀\nvless://uuid@tg.example.com:443?security=tls#TG\n👍"
	uris := parsers.ExtractConfigs(msg)

	if len(uris) != 1 {
		t.Fatalf("expected 1 URI from Telegram msg, got %d", len(uris))
	}
	if !strings.Contains(uris[0], "tg.example.com") {
		t.Errorf("unexpected URI: %s", uris[0])
	}
}

func TestExtract_TrailingPunctuation(t *testing.T) {
	text := `link1: vless://uuid@a.com:443?security=tls#A.
link2: ss://base64data@b.com:8388#B!
link3: trojan://pass@c.com:443#C)`

	uris := parsers.ExtractURIs(text)
	for _, u := range uris {
		last := u[len(u)-1]
		if last == '.' || last == '!' || last == ')' {
			t.Errorf("trailing punctuation not stripped: %s", u)
		}
	}
}

func TestExtract_HTMLEntities(t *testing.T) {
	text := "vless://uuid@html.com:443?security=tls&amp;sni=html.com#HTMLTest"
	uris := parsers.ExtractURIs(text)

	if len(uris) == 0 {
		t.Fatal("expected at least 1 URI after HTML entity cleanup")
	}
	if strings.Contains(uris[0], "&amp;") {
		t.Errorf("HTML entity &amp; was not decoded: %s", uris[0])
	}
}

func TestExtract_ZeroWidthChars(t *testing.T) {
	text := "vless://uuid\u200b@zw.com:443?security=tls#ZeroWidth"
	uris := parsers.ExtractURIs(text)

	if len(uris) == 0 {
		t.Log("zero-width chars broke extraction (expected behaviour if regex cant match)")
		return
	}
	for _, u := range uris {
		if strings.ContainsRune(u, '\u200b') {
			t.Errorf("zero-width space survived in URI: %s", u)
		}
	}
}

func TestExtract_DuplicatesRemoved(t *testing.T) {
	uri := "vless://uuid@dedup.com:443?security=tls#Dup"
	text := "first: " + uri + "\nsecond: " + uri + "\nthird: " + uri

	uris := parsers.ExtractURIs(text)
	if len(uris) != 1 {
		t.Errorf("expected 1 after dedup, got %d", len(uris))
	}
}

func TestExtract_PlainText_NoURIs(t *testing.T) {
	text := "این یک متن بدون هیچ لینکی است 🎭🎪 nothing to see here"
	uris := parsers.ExtractURIs(text)
	if len(uris) != 0 {
		t.Errorf("expected 0 URIs, got %d", len(uris))
	}
}

func TestExtract_EmptyString(t *testing.T) {
	uris := parsers.ExtractURIs("")
	if len(uris) != 0 {
		t.Errorf("expected 0 URIs from empty string, got %d", len(uris))
	}
}

func TestExtract_TruncatedURI_Rejected(t *testing.T) {
	text := "vless://uuid@trunc.com:443?security=tls..."
	uris := parsers.ExtractURIs(text)

	for _, u := range uris {
		if strings.HasSuffix(u, "...") {
			t.Errorf("truncated URI was not rejected: %s", u)
		}
	}
}
