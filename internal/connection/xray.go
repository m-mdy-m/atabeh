package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/m-mdy-m/atabeh/cmd/fs"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

const tag = "xray"

func NewXray() *Xray {
	baseDir := fs.BaseDir()
	logDir := filepath.Join(baseDir, "logs")
	os.MkdirAll(logDir, 0755)

	return &Xray{
		configPath: filepath.Join(baseDir, "xray-config.json"),
		logPath:    filepath.Join(logDir, "xray.log"),
	}
}

func (x *Xray) Start(configJSON []byte) error {
	if x.running {
		return fmt.Errorf("already running")
	}

	logger.Debugf(tag, "Writing config to %s", x.configPath)

	if err := os.WriteFile(x.configPath, configJSON, 0644); err != nil {
		logger.Errorf(tag, "Failed to write config: %v", err)
		return fmt.Errorf("write config: %w", err)
	}

	var test map[string]interface{}
	if err := json.Unmarshal(configJSON, &test); err != nil {
		logger.Errorf(tag, "Invalid JSON config: %v", err)
		return fmt.Errorf("invalid config: %w", err)
	}

	logger.Debugf(tag, "Config written (%d bytes)", len(configJSON))

	xrayBin, err := exec.LookPath("xray")
	if err != nil {
		logger.Errorf(tag, "xray not found: %v", err)
		return fmt.Errorf("xray not found (install from https://github.com/XTLS/Xray-core)")
	}

	logger.Debugf(tag, "Using xray: %s", xrayBin)

	testCmd := exec.Command(xrayBin, "test", "-c", x.configPath)
	if output, err := testCmd.CombinedOutput(); err != nil {
		logger.Errorf(tag, "Config test failed: %v", err)
		logger.Errorf(tag, "xray test output:\n%s", string(output))
		return fmt.Errorf("invalid config: %s", string(output))
	}

	logger.Debugf(tag, "Config validated successfully")

	logFile, err := os.OpenFile(x.logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Errorf(tag, "Failed to open log: %v", err)
		return fmt.Errorf("open log: %w", err)
	}

	x.ctx, x.cancel = context.WithCancel(context.Background())
	x.cmd = exec.CommandContext(x.ctx, xrayBin, "run", "-c", x.configPath)
	x.cmd.Stdout = logFile
	x.cmd.Stderr = logFile

	if err := x.cmd.Start(); err != nil {
		logFile.Close()
		logger.Errorf(tag, "Failed to start: %v", err)
		return fmt.Errorf("start: %w", err)
	}

	x.running = true
	logger.Infof(tag, "Started (PID: %d)", x.cmd.Process.Pid)

	if err := x.waitReady(); err != nil {
		x.Stop()

		logContent, _ := os.ReadFile(x.logPath)
		logger.Errorf(tag, "xray log:\n%s", string(logContent))

		return fmt.Errorf("failed to start: %w", err)
	}

	logger.Infof(tag, "Ready and listening")
	return nil
}

func (x *Xray) Stop() error {
	if !x.running {
		return nil
	}

	logger.Infof(tag, "Stopping...")

	if x.cancel != nil {
		x.cancel()
	}

	if x.cmd != nil && x.cmd.Process != nil {
		x.cmd.Process.Signal(os.Interrupt)

		done := make(chan error, 1)
		go func() {
			done <- x.cmd.Wait()
		}()

		select {
		case <-time.After(2 * time.Second):
			logger.Warnf(tag, "Force killing")
			x.cmd.Process.Kill()
		case <-done:
		}
	}

	x.running = false
	x.cmd = nil

	logger.Infof(tag, "Stopped")
	return nil
}

func (x *Xray) waitReady() error {
	logger.Debugf(tag, "Waiting for ready...")

	deadline := time.Now().Add(10 * time.Second)

	for time.Now().Before(deadline) {

		if x.cmd.Process == nil {
			return fmt.Errorf("process is nil")
		}

		if err := x.cmd.Process.Signal(os.Signal(nil)); err != nil {
			return fmt.Errorf("process died")
		}

		conn, err := net.DialTimeout("tcp", "127.0.0.1:10808", 100*time.Millisecond)
		if err == nil {
			conn.Close()
			logger.Debugf(tag, "SOCKS5 port is listening")
			return nil
		}

		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("timeout")
}

func (x *Xray) IsRunning() bool {
	if !x.running || x.cmd == nil || x.cmd.Process == nil {
		return false
	}

	return x.cmd.Process.Signal(os.Signal(nil)) == nil
}

func (x *Xray) GetStats() (uint64, uint64, error) {
	if !x.IsRunning() {
		return 0, 0, fmt.Errorf("not running")
	}

	resp, err := http.Get("http://127.0.0.1:10085/stats/query?pattern=")
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	var result struct {
		Stats []struct {
			Name  string `json:"name"`
			Value int64  `json:"value"`
		} `json:"stat"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, 0, err
	}

	var upload, download uint64
	for _, s := range result.Stats {
		if s.Name == "outbound>>>proxy>>>traffic>>>uplink" {
			upload = uint64(s.Value)
		} else if s.Name == "outbound>>>proxy>>>traffic>>>downlink" {
			download = uint64(s.Value)
		}
	}

	return upload, download, nil
}

func (x *Xray) GetProxyAddr() string {
	return "127.0.0.1:10808"
}

func (x *Xray) GetHTTPProxyAddr() string {
	return "127.0.0.1:10809"
}
