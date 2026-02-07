package command

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/connection"
	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) ConnectCommand() *cobra.Command {
	var (
		configID    int
		auto        bool
		realTime    bool
		systemProxy bool
		keepRunning bool
	)

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to VPN/proxy",
		Long: `Connect to a configuration.

MODES:
  --id N              Connect to specific config
  --auto              Auto-select best config (hourly updates)
  --auto --real-time  Auto-select with real-time monitoring (30s updates)

OPTIONS:
  --system-proxy      Enable system-wide proxy (requires root/admin)
  --keep              Keep connection active until Ctrl+C

EXAMPLES:
  # Connect to specific config
  atabeh connect --id 5 --keep

  # Auto-select best (recommended)
  atabeh connect --auto --keep

  # System-wide proxy (all apps use proxy)
  sudo atabeh connect --auto --system-proxy --keep

  # Real-time monitoring (for unstable networks)
  sudo atabeh connect --auto --real-time --system-proxy --keep`,
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			// Validate
			if auto && configID != 0 {
				return fmt.Errorf("use --auto OR --id, not both")
			}
			if !auto && configID == 0 {
				return fmt.Errorf("use --auto or --id <config-id>")
			}
			if realTime && !auto {
				return fmt.Errorf("--real-time requires --auto")
			}

			// Check root if system-proxy
			if systemProxy && !connection.IsRoot() {
				logger.Errorf("connect", "System proxy requires root/admin")
				fmt.Println("\nRun with:")
				fmt.Println("  Linux:   sudo atabeh connect ...")
				fmt.Println("  Windows: Run as Administrator")
				return fmt.Errorf("need root/admin for system proxy")
			}

			// Create manager
			mgr := connection.NewManager(repo)

			// Connect
			var err error
			if auto {
				logger.Infof("connect", "Auto mode: %s", getMode(realTime))
				err = mgr.ConnectAuto(realTime)
			} else {
				logger.Infof("connect", "Connecting to config %d...", configID)
				err = mgr.Connect(configID)
			}

			if err != nil {
				logger.Errorf("connect", "Failed: %v", err)
				return err
			}

			// Get status
			status := mgr.GetStatus()

			logger.Infof("connect", "✓ Connected!")
			logger.Infof("connect", "  Config:   %s", status.ConfigName)
			logger.Infof("connect", "  Server:   %s", status.Server)
			logger.Infof("connect", "  Protocol: %s", status.Protocol)

			if !systemProxy {
				logger.Infof("connect", "  SOCKS5:   127.0.0.1:10808")
				logger.Infof("connect", "  HTTP:     127.0.0.1:10809")
				logger.Infof("connect", "")
				logger.Infof("connect", "Configure your apps to use these proxies.")
			}

			// Enable system proxy
			if systemProxy {
				logger.Infof("connect", "")
				logger.Infof("connect", "Enabling system-wide proxy...")
				if err := mgr.EnableSystemProxy(); err != nil {
					logger.Errorf("connect", "Failed to enable proxy: %v", err)
				} else {
					logger.Infof("connect", "✓ System proxy enabled - all apps will use proxy")
				}
			}

			// Keep running
			if keepRunning {
				logger.Infof("connect", "")
				logger.Infof("connect", "Press Ctrl+C to disconnect...")
				logger.Infof("connect", "")

				// Show status updates
				go showStatus(mgr)

				// Wait for signal
				sig := make(chan os.Signal, 1)
				signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
				<-sig

				logger.Infof("connect", "")
				logger.Infof("connect", "Disconnecting...")

				if err := mgr.Disconnect(); err != nil {
					logger.Errorf("connect", "Disconnect error: %v", err)
				} else {
					logger.Infof("connect", "✓ Disconnected")
				}
			} else {
				logger.Infof("connect", "")
				logger.Infof("connect", "Connection active in background.")
				logger.Infof("connect", "Use 'atabeh status' to monitor.")
				logger.Infof("connect", "Use 'atabeh disconnect' to stop.")
			}

			return nil
		}),
	}

	cmd.Flags().IntVar(&configID, "id", 0, "config ID")
	cmd.Flags().BoolVar(&auto, "auto", false, "auto-select best")
	cmd.Flags().BoolVar(&realTime, "real-time", false, "real-time monitoring")
	cmd.Flags().BoolVar(&systemProxy, "system-proxy", false, "system-wide proxy")
	cmd.Flags().BoolVar(&keepRunning, "keep", false, "keep running")

	return cmd
}

func getMode(realTime bool) string {
	if realTime {
		return "real-time (30s updates)"
	}
	return "hourly updates"
}

func showStatus(mgr *connection.Manager) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		status := mgr.GetStatus()
		if !status.Connected {
			return
		}

		if status.Upload > 0 || status.Download > 0 {
			logger.Infof("connect", "↑ %s  ↓ %s  (speed: ↑ %s  ↓ %s)",
				connection.FormatBytes(status.Upload),
				connection.FormatBytes(status.Download),
				connection.FormatSpeed(status.UpSpeed),
				connection.FormatSpeed(status.DownSpeed))
		}
	}
}
