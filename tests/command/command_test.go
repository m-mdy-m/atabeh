package command_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/m-mdy-m/atabeh/cmd/command"
	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

func setupTestDB(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return dbPath, cleanup
}

func TestAddCommand(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name: "add single vless config",
			args: []string{
				"vless://12345678-1234-1234-1234-123456789012@example.com:443?type=tcp&security=tls#Test",
			},
			wantErr: false,
		},
		{
			name: "add multiple configs",
			args: []string{
				`vless://12345678-1234-1234-1234-123456789012@server1.com:443#Server1
vless://87654321-4321-4321-4321-210987654321@server2.com:443#Server2`,
			},
			wantErr: false,
		},
		{
			name: "add invalid config",
			args: []string{
				"invalid://not-a-config",
			},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := &command.CLI{DBPath: &dbPath}
			cmd := cli.AddCommand()

			cmd.Flags().Set("test-first", "false")
			cmd.Flags().Set("profile", "")

			err := cmd.RunE(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListCommand(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}
	repo := repository.NewFromDB(db)

	profileID, _ := repo.GetOrCreateProfile("Test Profile", "test", "manual")
	repo.InsertConfigBatch([]*common.NormalizedConfig{
		{
			Name:     "Test1",
			Protocol: "vless",
			Server:   "server1.com",
			Port:     443,
			UUID:     "12345678-1234-1234-1234-123456789012",
		},
		{
			Name:     "Test2",
			Protocol: "vmess",
			Server:   "server2.com",
			Port:     443,
			UUID:     "87654321-4321-4321-4321-210987654321",
		},
	}, profileID)
	db.Close()

	tests := []struct {
		name    string
		flags   map[string]string
		wantErr bool
	}{
		{
			name:    "list all configs",
			flags:   map[string]string{},
			wantErr: false,
		},
		{
			name: "list profiles",
			flags: map[string]string{
				"profiles": "true",
			},
			wantErr: false,
		},
		{
			name: "list by protocol",
			flags: map[string]string{
				"protocol": "vless",
			},
			wantErr: false,
		},
		{
			name: "list alive only",
			flags: map[string]string{
				"alive": "true",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := &command.CLI{DBPath: &dbPath}
			cmd := cli.ListCommand()

			for k, v := range tt.flags {
				cmd.Flags().Set(k, v)
			}

			err := cmd.RunE(cmd, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("ListCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveCommand(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}
	repo := repository.NewFromDB(db)

	profileID, _ := repo.GetOrCreateProfile("Test Profile", "test", "manual")
	id, _, _ := repo.InsertConfigOrSkip(&common.NormalizedConfig{
		Name:     "Test",
		Protocol: "vless",
		Server:   "test.com",
		Port:     443,
		UUID:     "12345678-1234-1234-1234-123456789012",
	}, profileID)
	db.Close()

	tests := []struct {
		name    string
		args    []string
		flags   map[string]string
		wantErr bool
	}{
		{
			name: "remove config with confirmation skip",
			args: []string{fmt.Sprintf("%d", id)},
			flags: map[string]string{
				"yes": "true",
			},
			wantErr: false,
		},
		{
			name: "remove non-existent config",
			args: []string{"9999"},
			flags: map[string]string{
				"yes": "true",
			},
			wantErr: true,
		},
		{
			name: "invalid id",
			args: []string{"not-a-number"},
			flags: map[string]string{
				"yes": "true",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := &command.CLI{DBPath: &dbPath}
			cmd := cli.RemoveCommand()

			for k, v := range tt.flags {
				cmd.Flags().Set(k, v)
			}

			err := cmd.RunE(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRankCommand(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}
	repo := repository.NewFromDB(db)

	profileID, _ := repo.GetOrCreateProfile("Test Profile", "test", "manual")
	repo.InsertConfigBatch([]*common.NormalizedConfig{
		{Name: "Fast", Protocol: "vless", Server: "fast.com", Port: 443, UUID: "12345678-1234-1234-1234-123456789012"},
		{Name: "Slow", Protocol: "vless", Server: "slow.com", Port: 443, UUID: "87654321-4321-4321-4321-210987654321"},
	}, profileID)
	db.Close()

	tests := []struct {
		name    string
		flags   map[string]string
		wantErr bool
	}{
		{
			name:    "rank all",
			flags:   map[string]string{},
			wantErr: false,
		},
		{
			name: "rank top 5",
			flags: map[string]string{
				"top": "5",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := &command.CLI{DBPath: &dbPath}
			cmd := cli.RankCommand()

			for k, v := range tt.flags {
				cmd.Flags().Set(k, v)
			}

			err := cmd.RunE(cmd, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("RankCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExportCommand(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}
	repo := repository.NewFromDB(db)

	profileID, _ := repo.GetOrCreateProfile("Test Profile", "test", "manual")
	repo.InsertConfigBatch([]*common.NormalizedConfig{
		{Name: "Test", Protocol: "vless", Server: "test.com", Port: 443, UUID: "12345678-1234-1234-1234-123456789012"},
	}, profileID)
	db.Close()

	tests := []struct {
		name    string
		flags   map[string]string
		wantErr bool
	}{
		{
			name: "export sing-box",
			flags: map[string]string{
				"profile": "1",
				"format":  "sing-box",
				"output":  filepath.Join(t.TempDir(), "test.json"),
			},
			wantErr: false,
		},
		{
			name: "missing profile",
			flags: map[string]string{
				"format": "sing-box",
			},
			wantErr: true,
		},
		{
			name: "non-existent profile",
			flags: map[string]string{
				"profile": "9999",
				"format":  "sing-box",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := &command.CLI{DBPath: &dbPath}
			cmd := cli.ExportCommand()

			for k, v := range tt.flags {
				cmd.Flags().Set(k, v)
			}

			err := cmd.RunE(cmd, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("ExportCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFullWorkflow(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	cli := &command.CLI{DBPath: &dbPath}

	addCmd := cli.AddCommand()
	err := addCmd.RunE(addCmd, []string{
		"vless://12345678-1234-1234-1234-123456789012@server1.com:443#Server1",
	})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	listCmd := cli.ListCommand()
	err = listCmd.RunE(listCmd, []string{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	listCmd2 := cli.ListCommand()
	listCmd2.Flags().Set("profiles", "true")
	err = listCmd2.RunE(listCmd2, []string{})
	if err != nil {
		t.Fatalf("List profiles failed: %v", err)
	}

	rankCmd := cli.RankCommand()
	err = rankCmd.RunE(rankCmd, []string{})
	if err != nil {
		t.Fatalf("Rank failed: %v", err)
	}

	exportCmd := cli.ExportCommand()
	exportCmd.Flags().Set("profile", "1")
	exportCmd.Flags().Set("format", "sing-box")
	exportCmd.Flags().Set("output", filepath.Join(t.TempDir(), "export.json"))
	err = exportCmd.RunE(exportCmd, []string{})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	removeCmd := cli.RemoveCommand()
	removeCmd.Flags().Set("all", "true")
	removeCmd.Flags().Set("yes", "true")
	err = removeCmd.RunE(removeCmd, []string{})
	if err != nil {
		t.Fatalf("Remove all failed: %v", err)
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("database errors", func(t *testing.T) {
		invalidPath := "/invalid/path/to/db.sqlite"
		cli := &command.CLI{DBPath: &invalidPath}

		cmd := cli.ListCommand()
		err := cmd.RunE(cmd, []string{})
		if err == nil {
			t.Error("Expected error with invalid DB path")
		}
	})

	t.Run("empty database operations", func(t *testing.T) {
		dbPath, cleanup := setupTestDB(t)
		defer cleanup()

		cli := &command.CLI{DBPath: &dbPath}

		listCmd := cli.ListCommand()
		err := listCmd.RunE(listCmd, []string{})
		if err != nil {
			t.Errorf("List empty db failed: %v", err)
		}

		rankCmd := cli.RankCommand()
		err = rankCmd.RunE(rankCmd, []string{})
		if err != nil {
			t.Errorf("Rank empty db failed: %v", err)
		}
	})
}
