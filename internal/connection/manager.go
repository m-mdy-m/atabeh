package connection

import (
	"context"
	"fmt"
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/exporter"
	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/internal/tester"
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

const managerTag = "manager"

// NewManager creates manager
func NewManager(repo *repository.Repo) *Manager {
	return &Manager{
		repo:  repo,
		xray:  NewXray(),
		proxy: NewSystemProxy(),
	}
}

// Connect connects to specific config
func (m *Manager) Connect(configID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connected {
		return fmt.Errorf("already connected")
	}

	logger.Infof(managerTag, "Getting config %d...", configID)

	config, err := m.repo.GetConfigByID(configID)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	return m.connect(config, false, false)
}

// ConnectAuto auto-selects best config
func (m *Manager) ConnectAuto(realTime bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connected {
		return fmt.Errorf("already connected")
	}

	logger.Infof(managerTag, "Auto-selecting best config...")

	// Get alive configs
	configs, err := m.repo.ListAliveConfigs()
	if err != nil {
		return fmt.Errorf("list configs: %w", err)
	}

	if len(configs) == 0 {
		return fmt.Errorf("no alive configs")
	}

	logger.Infof(managerTag, "Testing %d configs...", len(configs))

	// Test all
	best := m.selectBest(configs)
	if best == nil {
		return fmt.Errorf("no working config found")
	}

	logger.Infof(managerTag, "Selected: %s (ping: %dms)", best.Name, best.LastPing)

	return m.connect(best, true, realTime)
}

// connect internal connection logic
func (m *Manager) connect(config *storage.ConfigRow, autoMode, realTime bool) error {
	logger.Infof(managerTag, "Connecting to %s (%s:%d)...", config.Name, config.Server, config.Port)

	// Generate xray config
	xrayConfig, err := exporter.ToXray([]*storage.ConfigRow{config}, true)
	if err != nil {
		logger.Errorf(managerTag, "Export failed: %v", err)
		return fmt.Errorf("export: %w", err)
	}

	logger.Debugf(managerTag, "Generated config (%d bytes)", len(xrayConfig))

	// Start xray
	if err := m.xray.Start(xrayConfig); err != nil {
		logger.Errorf(managerTag, "Xray start failed: %v", err)
		return fmt.Errorf("start xray: %w", err)
	}

	// Update state
	m.connected = true
	m.config = config
	m.autoMode = autoMode
	m.realTime = realTime
	m.lastCheck = time.Now()

	// Start monitoring
	m.ctx, m.cancel = context.WithCancel(context.Background())
	go m.monitor()

	logger.Infof(managerTag, "Connected!")
	logger.Infof(managerTag, "SOCKS5: %s", m.xray.GetProxyAddr())
	logger.Infof(managerTag, "HTTP: %s", m.xray.GetHTTPProxyAddr())

	return nil
}

// Disconnect disconnects
func (m *Manager) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		return nil
	}

	logger.Infof(managerTag, "Disconnecting...")

	// Stop monitoring
	if m.cancel != nil {
		m.cancel()
	}

	// Stop xray
	m.xray.Stop()

	// Disable proxy
	m.proxy.Disable()

	m.connected = false
	m.config = nil

	logger.Infof(managerTag, "Disconnected")
	return nil
}

// EnableSystemProxy enables system proxy
func (m *Manager) EnableSystemProxy() error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}

	return m.proxy.Enable(m.xray.GetProxyAddr(), m.xray.GetHTTPProxyAddr())
}

// DisableSystemProxy disables system proxy
func (m *Manager) DisableSystemProxy() error {
	return m.proxy.Disable()
}

// GetStatus returns status
func (m *Manager) GetStatus() Status {
	m.mu.Lock()
	connected := m.connected
	config := m.config
	autoMode := m.autoMode
	realTime := m.realTime
	m.mu.Unlock()

	m.statsMu.RLock()
	defer m.statsMu.RUnlock()

	s := Status{
		Connected: connected,
		Upload:    m.uploadBytes,
		Download:  m.downloadBytes,
		UpSpeed:   m.uploadSpeed,
		DownSpeed: m.downloadSpeed,
	}

	if config != nil {
		s.ConfigName = config.Name
		s.Server = config.Server
		s.Protocol = string(config.Protocol)
	}

	s.AutoMode = autoMode
	s.RealTime = realTime

	return s
}

// monitor monitors connection
func (m *Manager) monitor() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.updateStats()
		}
	}
}

func (m *Manager) updateStats() {
	upload, download, err := m.xray.GetStats()
	if err != nil {
		return
	}

	m.statsMu.Lock()
	defer m.statsMu.Unlock()

	now := time.Now()
	elapsed := now.Sub(m.lastCheck).Seconds()

	if elapsed > 0 {
		uploadDiff := upload - m.prevUpload
		downloadDiff := download - m.prevDownload

		m.uploadSpeed = uint64(float64(uploadDiff) / elapsed)
		m.downloadSpeed = uint64(float64(downloadDiff) / elapsed)
	}

	m.uploadBytes = upload
	m.downloadBytes = download
	m.prevUpload = upload
	m.prevDownload = download
	m.lastCheck = now
}

func (m *Manager) selectBest(configs []*storage.ConfigRow) *storage.ConfigRow {
	// Convert to normalized
	norms := make([]*common.NormalizedConfig, len(configs))
	for i, c := range configs {
		norms[i] = &common.NormalizedConfig{
			Protocol:  c.Protocol,
			Server:    c.Server,
			Port:      c.Port,
			UUID:      c.UUID,
			Password:  c.Password,
			Method:    c.Method,
			Transport: c.Transport,
			Security:  c.Security,
		}
	}

	// Test
	cfg := tester.Config{
		Attempts:        3,
		Timeout:         5 * time.Second,
		ConcurrentTests: 20,
	}

	results := tester.TestAll(norms, cfg)

	// Find best
	var best *storage.ConfigRow
	var bestPing int64 = 999999

	for i, r := range results {
		if r.Reachable && r.AvgMs < bestPing {
			bestPing = r.AvgMs
			best = configs[i]
			best.LastPing = r.AvgMs
		}
	}

	return best
}

// FormatBytes formats bytes
func FormatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// FormatSpeed formats speed
func FormatSpeed(bytesPerSec uint64) string {
	return FormatBytes(bytesPerSec) + "/s"
}
