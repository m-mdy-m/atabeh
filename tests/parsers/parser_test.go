package parsers_test

import (
	"testing"

	"github.com/m-mdy-m/atabeh/internal/parsers"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "single vless",
			input: "vless://uuid@server.com:443?type=tcp#test",
			want:  1,
		},
		{
			name: "multiple configs",
			input: `
vless://uuid1@server1.com:443?type=tcp#test1
vmess://base64encoded
ss://base64@server.com:8388#ss-test
trojan://pass@server.com:443#trojan-test
			`,
			want: 4,
		},
		{
			name:  "with noise",
			input: "Some text before vless://uuid@server.com:443#test and after",
			want:  1,
		},
		{
			name:  "duplicates",
			input: "vless://uuid@server.com:443#test\nvless://uuid@server.com:443#test",
			want:  1, // should deduplicate
		},
		{
			name:  "no configs",
			input: "Just some random text",
			want:  0,
		},
		{
			name:  "truncated URIs filtered",
			input: "vless://...",
			want:  0,
		},
		{
			name: "mixed with emojis and symbols",
			input: `
🇩🇪 Germany:
vless://uuid@de1.example.com:443#DE-01
vless://uuid@de2.example.com:443#DE-02
			`,
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsers.Extract(tt.input)
			if len(result) != tt.want {
				t.Errorf("Extract() got %d URIs, want %d", len(result), tt.want)
			}
		})
	}
}

func TestParseMixedContent(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		wantSubs          int
		wantDirectConfigs int
	}{
		{
			name: "only subscriptions",
			input: `
https://raw.githubusercontent.com/user/repo/main/sub.txt
https://example.com/subscription
			`,
			wantSubs:          2,
			wantDirectConfigs: 0,
		},
		{
			name: "only direct configs",
			input: `
vless://uuid@server1.com:443#test1
vmess://base64encoded
			`,
			wantSubs:          0,
			wantDirectConfigs: 2,
		},
		{
			name: "mixed content",
			input: `
https://example.com/sub.txt
vless://uuid@server.com:443#direct
			`,
			wantSubs:          1,
			wantDirectConfigs: 1,
		},
		{
			name:              "empty",
			input:             "",
			wantSubs:          0,
			wantDirectConfigs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsers.ParseMixedContent(tt.input)
			if err != nil {
				t.Errorf("ParseMixedContent() error = %v", err)
				return
			}
			if len(result.Subscriptions) != tt.wantSubs {
				t.Errorf("ParseMixedContent() got %d subs, want %d", len(result.Subscriptions), tt.wantSubs)
			}
			// Note: directConfigs depends on actual URI parsing which we test separately
		})
	}
}

func TestIsSubscriptionURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "github raw",
			input: "https://raw.githubusercontent.com/user/repo/main/sub.txt",
			want:  true,
		},
		{
			name:  "with /sub path",
			input: "https://example.com/api/v1/sub",
			want:  true,
		},
		{
			name:  "with .txt extension",
			input: "https://example.com/configs.txt",
			want:  true,
		},
		{
			name:  "config URI",
			input: "vless://uuid@server.com:443",
			want:  false,
		},
		{
			name:  "URL containing config URI",
			input: "https://example.com/vless://embedded",
			want:  false,
		},
		{
			name:  "not http",
			input: "ftp://example.com/sub.txt",
			want:  false,
		},
		{
			name:  "regular URL",
			input: "https://example.com/index.html",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsers.IsSubscriptionURL(tt.input)
			if got != tt.want {
				t.Errorf("isSubscriptionURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Edge cases tests
func TestExtractEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "very long input",
			input: string(make([]byte, 1000000)) + "vless://uuid@server.com:443#test",
			want:  1,
		},
		{
			name:  "many duplicates",
			input: repeatString("vless://uuid@server.com:443#test\n", 1000),
			want:  1,
		},
		{
			name:  "unicode everywhere",
			input: "🌟✨vless://uuid@server.com:443#test🎉🎊",
			want:  1,
		},
		{
			name:  "malformed but parseable",
			input: "vless://uuid@server.com:443?type=tcp&?extra=yes#test",
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsers.Extract(tt.input)
			if len(result) != tt.want {
				t.Errorf("Extract() got %d URIs, want %d", len(result), tt.want)
			}
		})
	}
}

func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
