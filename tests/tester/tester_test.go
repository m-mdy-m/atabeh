package tester_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/tester"
)

func TestTagConnectionError(t *testing.T) {
	tests := []struct {
		name    string
		errMsg  string
		wantTag string
	}{
		{
			name:    "timeout",
			errMsg:  "dial tcp: i/o timeout",
			wantTag: "timeout",
		},
		{
			name:    "connection refused",
			errMsg:  "dial tcp: connection refused",
			wantTag: "refused",
		},
		{
			name:    "no route to host",
			errMsg:  "dial tcp: no route to host",
			wantTag: "no-route",
		},
		{
			name:    "DNS failure",
			errMsg:  "dial tcp: lookup example.com: no such host",
			wantTag: "dns-fail",
		},
		{
			name:    "connection reset (DPI)",
			errMsg:  "read tcp: connection reset by peer",
			wantTag: "dpi-reset",
		},
		{
			name:    "broken pipe",
			errMsg:  "write tcp: broken pipe",
			wantTag: "dpi-reset",
		},
		{
			name:    "generic network error",
			errMsg:  "some random network error",
			wantTag: "network-fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &net.OpError{Err: &net.DNSError{Err: tt.errMsg}}
			tag := tester.TagConnectionError(err)
			if tag != tt.wantTag {
				t.Errorf("tagConnectionError() = %q, want %q", tag, tt.wantTag)
			}
		})
	}
}

func TestIsSuspiciouslyFast(t *testing.T) {
	tests := []struct {
		name string
		ms   int64
		want bool
	}{
		{name: "very fast", ms: 1, want: true},
		{name: "fast", ms: 5, want: true},
		{name: "borderline", ms: 9, want: true},
		{name: "normal", ms: 10, want: false},
		{name: "slow", ms: 100, want: false},
		{name: "very slow", ms: 500, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tester.IsSuspiciouslyFast(tt.ms)
			if got != tt.want {
				t.Errorf("isSuspiciouslyFast(%d) = %v, want %v", tt.ms, got, tt.want)
			}
		})
	}
}

func TestMostCommon(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  string
	}{
		{
			name:  "clear winner",
			input: []string{"timeout", "timeout", "timeout", "dns-fail"},
			want:  "timeout",
		},
		{
			name:  "tie goes to first",
			input: []string{"timeout", "dns-fail"},
			want:  "timeout",
		},
		{
			name:  "single item",
			input: []string{"dpi-reset"},
			want:  "dpi-reset",
		},
		{
			name:  "empty",
			input: []string{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tester.MostCommon(tt.input)
			if got != tt.want {
				t.Errorf("mostCommon() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNeedsTLS(t *testing.T) {
	tests := []struct {
		name   string
		config *common.NormalizedConfig
		want   bool
	}{
		{
			name:   "TLS",
			config: &common.NormalizedConfig{Security: "tls"},
			want:   true,
		},
		{
			name:   "Reality",
			config: &common.NormalizedConfig{Security: "reality"},
			want:   true,
		},
		{
			name:   "TLS uppercase",
			config: &common.NormalizedConfig{Security: "TLS"},
			want:   true,
		},
		{
			name:   "None",
			config: &common.NormalizedConfig{Security: "none"},
			want:   false,
		},
		{
			name:   "Empty",
			config: &common.NormalizedConfig{Security: ""},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tester.NeedsTLS(tt.config)
			if got != tt.want {
				t.Errorf("needsTLS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSNI(t *testing.T) {
	tests := []struct {
		name   string
		config *common.NormalizedConfig
		want   string
	}{
		{
			name: "with SNI in extra",
			config: &common.NormalizedConfig{
				Server: "default.com",
				Extra:  map[string]string{"sni": "custom.com"},
			},
			want: "custom.com",
		},
		{
			name: "without SNI",
			config: &common.NormalizedConfig{
				Server: "server.com",
				Extra:  map[string]string{},
			},
			want: "server.com",
		},
		{
			name: "nil extra",
			config: &common.NormalizedConfig{
				Server: "server.com",
				Extra:  nil,
			},
			want: "server.com",
		},
		{
			name: "empty SNI",
			config: &common.NormalizedConfig{
				Server: "server.com",
				Extra:  map[string]string{"sni": ""},
			},
			want: "server.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tester.GetSNI(tt.config)
			if got != tt.want {
				t.Errorf("getSNI() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDialWithContextMock(t *testing.T) {

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	addr := listener.Addr().String()
	cfg := &common.NormalizedConfig{
		Server:   "127.0.0.1",
		Port:     listener.Addr().(*net.TCPAddr).Port,
		Security: "none",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, tag, err := tester.DialWithContext(ctx, "tcp", addr, cfg)
	if err != nil {
		t.Errorf("dialWithContext() error = %v", err)
		return
	}
	if conn != nil {
		conn.Close()
	}
	if tag != "" {
		t.Errorf("dialWithContext() tag = %q, want empty", tag)
	}
}

func TestTestAll(t *testing.T) {

	configs := []*common.NormalizedConfig{
		{
			Name:     "test1",
			Protocol: common.Vless,
			Server:   "192.0.2.1",
			Port:     443,
			Security: "none",
		},
		{
			Name:     "test2",
			Protocol: common.Vless,
			Server:   "192.0.2.2",
			Port:     443,
			Security: "none",
		},
	}

	cfg := tester.Config{
		Attempts:        1,
		Timeout:         10 * time.Millisecond,
		ConcurrentTests: 2,
		TestDelay:       0,
	}

	results := tester.TestAll(configs, cfg)

	if len(results) != len(configs) {
		t.Errorf("TestAll() returned %d results, want %d", len(results), len(configs))
	}

	for i, result := range results {
		if result.Config != configs[i] {
			t.Errorf("Result[%d].Config mismatch", i)
		}
		if result.Attempts != cfg.Attempts {
			t.Errorf("Result[%d].Attempts = %d, want %d", i, result.Attempts, cfg.Attempts)
		}

		if result.Reachable {
			t.Errorf("Result[%d].Reachable = true, want false (TEST-NET should be unreachable)", i)
		}
	}
}

func TestResultFields(t *testing.T) {
	cfg := &common.NormalizedConfig{
		Name:     "test",
		Protocol: common.Vless,
		Server:   "192.0.2.1",
		Port:     443,
	}

	result := &tester.Result{
		Config:         cfg,
		Reachable:      false,
		Attempts:       3,
		Successes:      0,
		AvgMs:          0,
		MinMs:          0,
		MaxMs:          0,
		StabilityScore: 0,
		FailureTag:     "timeout",
	}

	if result.Config != cfg {
		t.Error("Result.Config not set correctly")
	}
	if result.Reachable {
		t.Error("Result.Reachable should be false")
	}
	if result.FailureTag != "timeout" {
		t.Errorf("Result.FailureTag = %q, want %q", result.FailureTag, "timeout")
	}
}

func BenchmarkTagConnectionError(b *testing.B) {
	err := &net.OpError{Err: &net.DNSError{Err: "connection reset by peer"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tester.TagConnectionError(err)
	}
}

func BenchmarkNeedsTLS(b *testing.B) {
	cfg := &common.NormalizedConfig{Security: "tls"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tester.NeedsTLS(cfg)
	}
}

func BenchmarkGetSNI(b *testing.B) {
	cfg := &common.NormalizedConfig{
		Server: "server.com",
		Extra:  map[string]string{"sni": "custom.com"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tester.GetSNI(cfg)
	}
}

func TestConcurrentTests(t *testing.T) {
	configs := make([]*common.NormalizedConfig, 100)
	for i := 0; i < 100; i++ {
		configs[i] = &common.NormalizedConfig{
			Name:     "test",
			Protocol: common.Vless,
			Server:   "192.0.2.1",
			Port:     443,
		}
	}

	cfg := tester.Config{
		Attempts:        1,
		Timeout:         10 * time.Millisecond,
		ConcurrentTests: 50,
		TestDelay:       0,
	}

	results := tester.TestAll(configs, cfg)

	if len(results) != len(configs) {
		t.Errorf("Got %d results, want %d", len(results), len(configs))
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty config list", func(t *testing.T) {
		results := tester.TestAll([]*common.NormalizedConfig{}, tester.Config{})
		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
	})

	t.Run("zero attempts", func(t *testing.T) {
		cfg := &common.NormalizedConfig{
			Server: "test.com",
			Port:   443,
		}
		result := tester.Test(cfg, tester.Config{Attempts: 0, Timeout: time.Second})
		if result.Attempts != 0 {
			t.Errorf("Expected 0 attempts, got %d", result.Attempts)
		}
	})

	t.Run("very long timeout", func(t *testing.T) {
		cfg := &common.NormalizedConfig{
			Server: "192.0.2.1",
			Port:   443,
		}
		result := tester.Test(cfg, tester.Config{
			Attempts: 1,
			Timeout:  1 * time.Hour,
		})
		if result.Reachable {
			t.Error("Should not be reachable")
		}
	})
}
