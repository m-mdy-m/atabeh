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
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	done := make(chan struct{})
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-done:
				default:
				}
				return
			}
			conn.Close()
		}
	}()
	return l.Addr().(*net.TCPAddr).Port, func() { close(done); l.Close() }
}

func startSlowServer(t *testing.T, delay time.Duration) (int, func()) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	done := make(chan struct{})
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-done:
				default:
				}
				return
			}
			time.Sleep(delay)
			conn.Close()
		}
	}()
	return l.Addr().(*net.TCPAddr).Port, func() { close(done); l.Close() }
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	p := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return p
}

func cfg(name string, port int) *common.NormalizedConfig {
	return &common.NormalizedConfig{
		Name:     name,
		Protocol: common.Vless,
		Server:   "127.0.0.1",
		Port:     port,
	}
}

func quickCfg() tester.Config {
	return tester.Config{
		Attempts:        3,
		Timeout:         2 * time.Second,
		ConcurrentTests: 5,
		TestDelay:       50 * time.Millisecond,
	}
}

func TestTester_ReachableServer_AllSucceed(t *testing.T) {
	port, stop := startTCPServer(t)
	defer stop()

	result := tester.Test(cfg("Alive", port), quickCfg())

	if !result.Reachable {
		t.Fatal("expected reachable")
	}
	if result.Successes != 3 {
		t.Errorf("successes: got %d, want 3", result.Successes)
	}
	if result.LossPercent != 0 {
		t.Errorf("loss: got %d%%, want 0", result.LossPercent)
	}
}

func TestTester_UnreachableServer_AllFail(t *testing.T) {
	c := &common.NormalizedConfig{
		Name:   "Dead",
		Server: "192.0.2.1",
		Port:   12345,
	}
	tc := tester.Config{
		Attempts: 2,
		Timeout:  500 * time.Millisecond,
	}

	result := tester.Test(c, tc)

	if result.Reachable {
		t.Fatal("expected unreachable")
	}
	if result.Successes != 0 {
		t.Errorf("successes: got %d, want 0", result.Successes)
	}
	if result.LossPercent != 100 {
		t.Errorf("loss: got %d%%, want 100", result.LossPercent)
	}
}

func TestTester_ConnectionRefused(t *testing.T) {
	port := freePort(t)

	result := tester.Test(cfg("Refused", port), tester.Config{
		Attempts: 2,
		Timeout:  1 * time.Second,
	})

	if result.Reachable {
		t.Errorf("connection-refused port should be unreachable")
	}
	if result.LossPercent != 100 {
		t.Errorf("loss: got %d%%, want 100", result.LossPercent)
	}
}

func TestTester_LatencyStats(t *testing.T) {
	port, stop := startTCPServer(t)
	defer stop()

	result := tester.Test(cfg("Latency", port), tester.Config{
		Attempts:  5,
		Timeout:   2 * time.Second,
		TestDelay: 10 * time.Millisecond,
	})

	if !result.Reachable {
		t.Fatal("expected reachable")
	}
	if result.AvgMs > 50 {
		t.Errorf("localhost avg latency too high: %d ms", result.AvgMs)
	}
	if result.MinMs > result.AvgMs {
		t.Errorf("min (%d) > avg (%d)", result.MinMs, result.AvgMs)
	}
	if result.MaxMs < result.AvgMs {
		t.Errorf("max (%d) < avg (%d)", result.MaxMs, result.AvgMs)
	}
	if result.MinMs > result.MaxMs {
		t.Errorf("min (%d) > max (%d)", result.MinMs, result.MaxMs)
	}
}

func TestTester_TestAll_MixedResults(t *testing.T) {
	alivePort, stop := startTCPServer(t)
	defer stop()
	deadPort := freePort(t)

	configs := []*common.NormalizedConfig{
		cfg("Alive1", alivePort),
		cfg("Dead1", deadPort),
		cfg("Alive2", alivePort),
		cfg("Dead2", deadPort),
	}

	results := tester.TestAll(configs, tester.Config{
		Attempts:        1,
		Timeout:         500 * time.Millisecond,
		ConcurrentTests: 4,
	})

	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
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
	if results[3].Reachable {
		t.Error("Dead2 should not be reachable")
	}
}

func TestTester_ConcurrencyStress(t *testing.T) {
	port, stop := startTCPServer(t)
	defer stop()

	n := 20
	configs := make([]*common.NormalizedConfig, n)
	for i := range configs {
		configs[i] = cfg("stress", port)
	}

	results := tester.TestAll(configs, tester.Config{
		Attempts:        1,
		Timeout:         2 * time.Second,
		ConcurrentTests: 10,
	})

	if len(results) != n {
		t.Fatalf("expected %d results, got %d", n, len(results))
	}
	for i, r := range results {
		if !r.Reachable {
			t.Errorf("result[%d] should be reachable", i)
		}
	}
}

func TestTester_RankResults_Order(t *testing.T) {
	results := []*common.PingResult{
		{Config: &common.NormalizedConfig{Name: "Dead"}, Reachable: false, LossPercent: 100},
		{Config: &common.NormalizedConfig{Name: "Slow"}, Reachable: true, AvgMs: 200, LossPercent: 0},
		{Config: &common.NormalizedConfig{Name: "Fast"}, Reachable: true, AvgMs: 50, LossPercent: 0},
		{Config: &common.NormalizedConfig{Name: "Mid"}, Reachable: true, AvgMs: 100, LossPercent: 33},
	}

	ranked := tester.RankResults(results)

	reachableEnded := false
	for _, r := range ranked {
		if !r.Reachable {
			reachableEnded = true
		}
		if reachableEnded && r.Reachable {
			t.Error("unreachable entry appeared before a reachable one")
		}
	}

	if ranked[0].Config.Name != "Fast" {
		t.Errorf("rank[0]: got %q, want Fast", ranked[0].Config.Name)
	}
	if ranked[1].Config.Name != "Slow" {
		t.Errorf("rank[1]: got %q, want Slow", ranked[1].Config.Name)
	}
	if ranked[2].Config.Name != "Mid" {
		t.Errorf("rank[2]: got %q, want Mid", ranked[2].Config.Name)
	}
	if ranked[3].Config.Name != "Dead" {
		t.Errorf("rank[3]: got %q, want Dead", ranked[3].Config.Name)
	}
}

func TestTester_RankResults_AllDead(t *testing.T) {
	results := []*common.PingResult{
		{Config: &common.NormalizedConfig{Name: "A"}, Reachable: false},
		{Config: &common.NormalizedConfig{Name: "B"}, Reachable: false},
	}
	ranked := tester.RankResults(results)
	if len(ranked) != 2 {
		t.Fatalf("expected 2, got %d", len(ranked))
	}
}

func TestTester_RankResults_AllAlive_SameLoss(t *testing.T) {
	results := []*common.PingResult{
		{Config: &common.NormalizedConfig{Name: "Slow"}, Reachable: true, AvgMs: 300, LossPercent: 0},
		{Config: &common.NormalizedConfig{Name: "Fast"}, Reachable: true, AvgMs: 10, LossPercent: 0},
		{Config: &common.NormalizedConfig{Name: "Mid"}, Reachable: true, AvgMs: 100, LossPercent: 0},
	}
	ranked := tester.RankResults(results)
	if ranked[0].Config.Name != "Fast" {
		t.Errorf("rank[0]: got %q, want Fast", ranked[0].Config.Name)
	}
	if ranked[2].Config.Name != "Slow" {
		t.Errorf("rank[2]: got %q, want Slow", ranked[2].Config.Name)
	}
}

func TestTester_DefaultConfig_Values(t *testing.T) {
	dc := tester.DefaultConfig()
	if dc.Attempts != 3 {
		t.Errorf("Attempts: got %d, want 3", dc.Attempts)
	}
	if dc.Timeout != 5*time.Second {
		t.Errorf("Timeout: got %v, want 5s", dc.Timeout)
	}
	if dc.ConcurrentTests != 10 {
		t.Errorf("ConcurrentTests: got %d, want 10", dc.ConcurrentTests)
	}
}

func TestTester_SlowServer_StillReachable(t *testing.T) {

	port, stop := startSlowServer(t, 200*time.Millisecond)
	defer stop()

	result := tester.Test(cfg("Slow", port), tester.Config{
		Attempts: 2,
		Timeout:  2 * time.Second,
	})

	if !result.Reachable {
		t.Error("slow server (200 ms) should still be reachable within 2 s timeout")
	}
}

func TestTester_TimeoutBeforeConnect(t *testing.T) {

	port, stop := startSlowServer(t, 5*time.Second)
	defer stop()

	_ = port

	c := &common.NormalizedConfig{
		Name:   "Timeout",
		Server: "192.0.2.1",
		Port:   9999,
	}
	result := tester.Test(c, tester.Config{
		Attempts: 1,
		Timeout:  100 * time.Millisecond,
	})

	if result.Reachable {
		t.Error("should timeout against TEST-NET-1")
	}
}

func TestTester_TestAll_NoIndexAlias(t *testing.T) {
	alivePort, stop := startTCPServer(t)
	defer stop()

	configs := make([]*common.NormalizedConfig, 10)
	deadPort := freePort(t)
	for i := range configs {
		if i%2 == 0 {
			configs[i] = cfg("alive", alivePort)
		} else {
			configs[i] = cfg("dead", deadPort)
		}
	}

	results := tester.TestAll(configs, tester.Config{
		Attempts:        1,
		Timeout:         500 * time.Millisecond,
		ConcurrentTests: 10,
	})

	for i, r := range results {
		wantAlive := i%2 == 0
		if r.Reachable != wantAlive {
			t.Errorf("results[%d]: reachable=%v, want %v", i, r.Reachable, wantAlive)
		}
	}
}
