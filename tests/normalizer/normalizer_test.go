package normalizer_test

import (
	"testing"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/normalizer"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *common.RawConfig
		wantErr bool
	}{
		{
			name: "valid vless",
			config: &common.RawConfig{
				Protocol: common.Vless,
				Server:   "example.com",
				Port:     443,
				UUID:     "12345678-1234-1234-1234-123456789012",
			},
			wantErr: false,
		},
		{
			name: "invalid vless - missing UUID",
			config: &common.RawConfig{
				Protocol: common.Vless,
				Server:   "example.com",
				Port:     443,
			},
			wantErr: true,
		},
		{
			name: "invalid vless - bad UUID format",
			config: &common.RawConfig{
				Protocol: common.Vless,
				Server:   "example.com",
				Port:     443,
				UUID:     "not-a-uuid",
			},
			wantErr: true,
		},
		{
			name: "valid shadowsocks",
			config: &common.RawConfig{
				Protocol: common.Shadowsocks,
				Server:   "1.2.3.4",
				Port:     8388,
				Password: "secret",
				Method:   "aes-256-gcm",
			},
			wantErr: false,
		},
		{
			name: "invalid shadowsocks - unsupported method",
			config: &common.RawConfig{
				Protocol: common.Shadowsocks,
				Server:   "1.2.3.4",
				Port:     8388,
				Password: "secret",
				Method:   "rc4-md5",
			},
			wantErr: true,
		},
		{
			name: "invalid - private IP",
			config: &common.RawConfig{
				Protocol: common.Vless,
				Server:   "192.168.1.1",
				Port:     443,
				UUID:     "12345678-1234-1234-1234-123456789012",
			},
			wantErr: true,
		},
		{
			name: "invalid - localhost",
			config: &common.RawConfig{
				Protocol: common.Vless,
				Server:   "127.0.0.1",
				Port:     443,
				UUID:     "12345678-1234-1234-1234-123456789012",
			},
			wantErr: true,
		},
		{
			name: "invalid - port out of range",
			config: &common.RawConfig{
				Protocol: common.Vless,
				Server:   "example.com",
				Port:     99999,
				UUID:     "12345678-1234-1234-1234-123456789012",
			},
			wantErr: true,
		},
		{
			name: "invalid - empty server",
			config: &common.RawConfig{
				Protocol: common.Vless,
				Server:   "",
				Port:     443,
				UUID:     "12345678-1234-1234-1234-123456789012",
			},
			wantErr: true,
		},
		{
			name: "valid trojan",
			config: &common.RawConfig{
				Protocol: common.Trojan,
				Server:   "example.com",
				Port:     443,
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "invalid trojan - missing password",
			config: &common.RawConfig{
				Protocol: common.Trojan,
				Server:   "example.com",
				Port:     443,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := normalizer.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCleanName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "with emojis",
			input: "Aleph ❤️🤍💚 FREE2CONFIG",
			want:  "Aleph FREE2CONFIG",
		},
		{
			name:  "with symbols",
			input: "【Server】🇩🇪 Germany-01",
			want:  "Server Germany-01",
		},
		{
			name:  "with location prefix",
			input: "🇺🇸42-New York",
			want:  "New York",
		},
		{
			name:  "URL encoded",
			input: "Server%20%231",
			want:  "Server #1",
		},
		{
			name:  "multiple spaces and dashes",
			input: "Server   ---   Name",
			want:  "Server Name",
		},
		{
			name:  "complex case",
			input: "🇩🇪42-【VIP】@channel Server★01",
			want:  "VIP @channel Server01",
		},
		{
			name:  "empty after cleaning",
			input: "🇩🇪@@@",
			want:  "",
		},
		{
			name:  "already clean",
			input: "Clean Server Name",
			want:  "Clean Server Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizer.CleanName(tt.input)
			if got != tt.want {
				t.Errorf("CleanName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeduplicate(t *testing.T) {
	tests := []struct {
		name   string
		input  []*common.NormalizedConfig
		expect int
	}{
		{
			name: "no duplicates",
			input: []*common.NormalizedConfig{
				{Protocol: common.Vless, Server: "server1.com", Port: 443, UUID: "uuid1"},
				{Protocol: common.Vless, Server: "server2.com", Port: 443, UUID: "uuid2"},
			},
			expect: 2,
		},
		{
			name: "exact duplicates",
			input: []*common.NormalizedConfig{
				{Protocol: common.Vless, Server: "server1.com", Port: 443, UUID: "uuid1", Transport: common.TCP},
				{Protocol: common.Vless, Server: "server1.com", Port: 443, UUID: "uuid1", Transport: common.TCP},
				{Protocol: common.Vless, Server: "server1.com", Port: 443, UUID: "uuid1", Transport: common.TCP},
			},
			expect: 1,
		},
		{
			name: "different names same config",
			input: []*common.NormalizedConfig{
				{Name: "Server A", Protocol: common.Vless, Server: "server1.com", Port: 443, UUID: "uuid1", Transport: common.TCP},
				{Name: "Server B", Protocol: common.Vless, Server: "server1.com", Port: 443, UUID: "uuid1", Transport: common.TCP},
			},
			expect: 1,
		},
		{
			name: "different transport not duplicate",
			input: []*common.NormalizedConfig{
				{Protocol: common.Vless, Server: "server1.com", Port: 443, UUID: "uuid1", Transport: common.TCP},
				{Protocol: common.Vless, Server: "server1.com", Port: 443, UUID: "uuid1", Transport: common.WS},
			},
			expect: 2,
		},
		{
			name: "shadowsocks dedup by password+method",
			input: []*common.NormalizedConfig{
				{Protocol: common.Shadowsocks, Server: "1.2.3.4", Port: 8388, Password: "pass1", Method: "aes-256-gcm"},
				{Protocol: common.Shadowsocks, Server: "1.2.3.4", Port: 8388, Password: "pass1", Method: "aes-256-gcm"},
				{Protocol: common.Shadowsocks, Server: "1.2.3.4", Port: 8388, Password: "pass2", Method: "aes-256-gcm"},
			},
			expect: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizer.Deduplicate(tt.input)
			if len(result) != tt.expect {
				t.Errorf("Deduplicate() got %d configs, want %d", len(result), tt.expect)
			}
		})
	}
}

func TestExtractProfileName(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "from URL fragment",
			source: "https://example.com/sub#MyProfile",
			want:   "MyProfile",
		},
		{
			name:   "from filename",
			source: "https://example.com/configs/iran-servers.txt",
			want:   "Iran-servers",
		},
		{
			name:   "from domain",
			source: "https://myservice.example.com/sub",
			want:   "Sub",
		},
		{
			name:   "from config URI fragment",
			source: "vless://uuid@server:443#ProfileName",
			want:   "ProfileName",
		},
		{
			name:   "complex URL",
			source: "https://raw.githubusercontent.com/user/repo/main/configs.txt",
			want:   "Configs",
		},
		{
			name:   "fallback",
			source: "unknown-format",
			want:   "Configs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizer.ExtractProfileName(tt.source)
			if got != tt.want {
				t.Errorf("ExtractProfileName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidate(b *testing.B) {
	cfg := &common.RawConfig{
		Protocol: common.Vless,
		Server:   "example.com",
		Port:     443,
		UUID:     "12345678-1234-1234-1234-123456789012",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizer.Validate(cfg)
	}
}

func BenchmarkCleanName(b *testing.B) {
	name := "🇩🇪42-【VIP】@channel Server★01 FREE2CONFIG"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizer.CleanName(name)
	}
}

func BenchmarkDeduplicate(b *testing.B) {
	configs := make([]*common.NormalizedConfig, 1000)
	for i := 0; i < 1000; i++ {
		configs[i] = &common.NormalizedConfig{
			Protocol:  common.Vless,
			Server:    "server.com",
			Port:      443,
			UUID:      "12345678-1234-1234-1234-123456789012",
			Transport: common.TCP,
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizer.Deduplicate(configs)
	}
}
