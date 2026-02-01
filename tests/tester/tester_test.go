package tester_test

import (
	"net"
	"testing"
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/tester"
)

func startTCPServer(t *testing.T) (int, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test TCP server: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	// Accept connections in background — just accept and close (simulate a live server)
	done := make(chan struct{})
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					return
				}
			}
			conn.Close()
		}
	}()

	stop := func() {
		close(done)
		listener.Close()
	}

	return port, stop
}

func TestTester_ReachableServer(t *testing.T) {
	port, stop := startTCPServer(t)
	defer stop()

	cfg := &common.NormalizedConfig{
		Name:     "LocalTest",
		Protocol: common.Vless,
		Server:   "127.0.0.1",
		Port:     port,
	}

	testCfg := tester.Config{
		Attempts: 3,
		Timeout:  2 * time.Second,
	}

	result := tester.Test(cfg, testCfg)

	if !result.Reachable {
		t.Error("expected server to be reachable")
	}
	if result.Successes != 3 {
		t.Errorf("expected 3 successes, got %d", result.Successes)
	}
	if result.LossPercent != 0 {
		t.Errorf("expected 0%% loss, got %d%%", result.LossPercent)
	}
	if result.AvgMs < 0 {
		t.Errorf("expected positive avg latency")
	}
}

func TestTester_UnreachableServer(t *testing.T) {
	cfg := &common.NormalizedConfig{
		Name:     "DeadServer",
		Protocol: common.Vless,
		Server:   "192.0.2.1", // TEST-NET-1, guaranteed unreachable
		Port:     12345,
	}

	testCfg := tester.Config{
		Attempts: 2,
		Timeout:  500 * time.Millisecond, // short timeout for test speed
	}

	result := tester.Test(cfg, testCfg)

	if result.Reachable {
		t.Error("expected server to be unreachable")
	}
	if result.Successes != 0 {
		t.Errorf("expected 0 successes, got %d", result.Successes)
	}
	if result.LossPercent != 100 {
		t.Errorf("expected 100%% loss, got %d%%", result.LossPercent)
	}
}

func TestTester_LatencyMeasurement(t *testing.T) {
	port, stop := startTCPServer(t)
	defer stop()

	cfg := &common.NormalizedConfig{
		Name:   "LatencyTest",
		Server: "127.0.0.1",
		Port:   port,
	}

	testCfg := tester.Config{
		Attempts: 5,
		Timeout:  2 * time.Second,
	}

	result := tester.Test(cfg, testCfg)

	// localhost should be < 100ms
	if result.AvgMs > 100 {
		t.Errorf("localhost avg latency too high: %d ms", result.AvgMs)
	}
	if result.MinMs > result.AvgMs {
		t.Errorf("min (%d) should be <= avg (%d)", result.MinMs, result.AvgMs)
	}
	if result.MaxMs < result.AvgMs {
		t.Errorf("max (%d) should be >= avg (%d)", result.MaxMs, result.AvgMs)
	}
}

func TestTester_TestAll(t *testing.T) {
	port, stop := startTCPServer(t)
	defer stop()

	configs := []*common.NormalizedConfig{
		{Name: "Alive1", Server: "127.0.0.1", Port: port},
		{Name: "Dead1", Server: "192.0.2.1", Port: 11111},
		{Name: "Alive2", Server: "127.0.0.1", Port: port},
	}

	testCfg := tester.Config{
		Attempts: 1,
		Timeout:  500 * time.Millisecond,
	}

	results := tester.TestAll(configs, testCfg)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if !results[0].Reachable {
		t.Error("Alive1 should be reachable")
	}
	if results[1].Reachable {
		t.Error("Dead1 should not be reachable")
	}
	if !results[2].Reachable {
		t.Error("Alive2 should be reachable")
	}
}

func TestTester_DefaultConfig(t *testing.T) {
	cfg := tester.DefaultConfig()

	if cfg.Attempts != 3 {
		t.Errorf("expected default 3 attempts, got %d", cfg.Attempts)
	}
	if cfg.Timeout != 5*time.Second {
		t.Errorf("expected default 5s timeout, got %v", cfg.Timeout)
	}
}

// TestTester_RefusedPort tests against a port with nothing listening (immediate refuse)
func TestTester_RefusedPort(t *testing.T) {
	// Find a free port, don't listen on it -> connection refused
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close() // close immediately -> port is free but nothing listening

	cfg := &common.NormalizedConfig{
		Name:   "RefusedPort",
		Server: "127.0.0.1",
		Port:   port,
	}

	testCfg := tester.Config{
		Attempts: 2,
		Timeout:  1 * time.Second,
	}

	result := tester.Test(cfg, testCfg)

	if result.Reachable {
		t.Errorf("expected connection refused on port %d", port)
	}
}
