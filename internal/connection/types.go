package connection

import (
	"context"
	"os/exec"
	"sync"
	"time"

	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

type Xray struct {
	cmd        *exec.Cmd
	configPath string
	logPath    string
	running    bool
	ctx        context.Context
	cancel     context.CancelFunc
}

type SystemProxy struct {
	enabled bool
}

type Manager struct {
	repo  *repository.Repo
	xray  *Xray
	proxy *SystemProxy

	mu        sync.Mutex
	connected bool
	config    *storage.ConfigRow
	autoMode  bool
	realTime  bool

	// Stats
	statsMu       sync.RWMutex
	uploadBytes   uint64
	downloadBytes uint64
	uploadSpeed   uint64
	downloadSpeed uint64
	lastCheck     time.Time
	prevUpload    uint64
	prevDownload  uint64

	ctx    context.Context
	cancel context.CancelFunc
}

// Status represents connection status
type Status struct {
	Connected  bool
	ConfigName string
	Server     string
	Protocol   string
	AutoMode   bool
	RealTime   bool
	Upload     uint64
	Download   uint64
	UpSpeed    uint64
	DownSpeed  uint64
}
