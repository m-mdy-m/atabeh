package command

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/m-mdy-m/atabeh/internal/connection"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

func (c *CLI) StatusCommand() *cobra.Command {
	var watch bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show connection status",
		RunE: c.WrapRepo(func(repo *repository.Repo, cmd *cobra.Command, args []string) error {
			mgr := connection.NewManager(repo)

			if watch {
				return watchStatus(mgr)
			}

			return printStatus(mgr)
		}),
	}

	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch mode")

	return cmd
}

func printStatus(mgr *connection.Manager) error {
	status := mgr.GetStatus()

	if !status.Connected {
		fmt.Println("\nStatus: Not connected")
		fmt.Println("\nUse 'atabeh connect' to connect")
		return nil
	}

	fmt.Println("\n✓ Connected")
	fmt.Println()
	fmt.Printf("  Config:   %s\n", status.ConfigName)
	fmt.Printf("  Server:   %s\n", status.Server)
	fmt.Printf("  Protocol: %s\n", status.Protocol)

	if status.AutoMode {
		mode := "Hourly updates"
		if status.RealTime {
			mode = "Real-time (30s)"
		}
		fmt.Printf("  Mode:     Auto (%s)\n", mode)
	} else {
		fmt.Printf("  Mode:     Manual\n")
	}

	fmt.Println()
	fmt.Printf("  Upload:   %s\n", connection.FormatBytes(status.Upload))
	fmt.Printf("  Download: %s\n", connection.FormatBytes(status.Download))
	fmt.Printf("  Total:    %s\n", connection.FormatBytes(status.Upload+status.Download))
	fmt.Println()

	if status.UpSpeed > 0 || status.DownSpeed > 0 {
		fmt.Printf("  Speed ↑:  %s\n", connection.FormatSpeed(status.UpSpeed))
		fmt.Printf("  Speed ↓:  %s\n", connection.FormatSpeed(status.DownSpeed))
		fmt.Println()
	}

	return nil
}

func watchStatus(mgr *connection.Manager) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		// Clear screen
		fmt.Print("\033[H\033[2J")

		printStatus(mgr)

		fmt.Println("Press Ctrl+C to stop watching...")

		<-ticker.C
	}
}
