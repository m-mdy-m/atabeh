package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/normalizer"
	"github.com/m-mdy-m/atabeh/internal/parsers"
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/repository"
)

func TestEdgeCase_VeryLongNames(t *testing.T) {
	longName := string(make([]byte, 1000))
	for i := range longName {
		longName = longName[:i] + "A"
	}

	cfg := &common.RawConfig{
		Protocol: common.Vless,
		Name:     longName,
		Server:   "example.com",
		Port:     443,
		UUID:     "12345678-1234-1234-1234-123456789012",
	}

	normalized, err := normalizer.Normalize([]*common.RawConfig{cfg})
	if err != nil {
		t.Fatalf("Normalize failed: %v", err)
	}

	if len(normalized) != 1 {
		t.Errorf("Expected 1 config, got %d", len(normalized))
	}

	if len(normalized[0].Name) > 500 {
		t.Errorf("Name too long: %d characters", len(normalized[0].Name))
	}
}

func TestEdgeCase_UnicodeEverywhere(t *testing.T) {
	text := `
🇩🇪🇺🇸🇬🇧🇫🇷🇯🇵🇨🇳
vless://12345678-1234-1234-1234-123456789012@example.com:443#🌟✨Test💫⭐
vmess://YmFzZTY0ZW5jb2RlZA==#مرورگر🔥
ss://base64@server.com:8388#😀😃😄😁😆
`

	uris := parsers.Extract(text)
	if len(uris) < 3 {
		t.Errorf("Expected at least 3 URIs, got %d", len(uris))
	}

	configs, err := parsers.ParseURIs(uris)
	if err != nil {
		t.Fatalf("ParseURIs failed: %v", err)
	}

	if len(configs) == 0 {
		t.Error("No configs parsed from unicode-heavy text")
	}
}

func TestEdgeCase_MaliciousInput(t *testing.T) {
	malicious := []string{
		"vless://'; DROP TABLE configs; --@evil.com:443#hack",
		"vless://<script>alert(1)</script>@xss.com:443#xss",
		"vless://../../../etc/passwd@path.com:443#traversal",
		string(make([]byte, 10000000)),
	}

	for i, input := range malicious {
		t.Run("malicious_"+string(rune(i)), func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panicked on malicious input: %v", r)
				}
			}()

			parsers.Extract(input)
			normalizer.CleanName(input)
		})
	}
}

func TestEdgeCase_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "concurrent.db")

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}
	defer db.Close()

	repo := repository.NewFromDB(db)
	profileID, _ := repo.GetOrCreateProfile("Test", "test", "manual")

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			defer func() { done <- true }()

			for j := 0; j < 10; j++ {
				cfg := &common.NormalizedConfig{
					Name:     fmt.Sprintf("Config-%d-%d", n, j),
					Protocol: common.Vless,
					Server:   fmt.Sprintf("server%d.com", n),
					Port:     443,
					UUID:     "12345678-1234-1234-1234-123456789012",
				}
				repo.InsertConfigOrSkip(cfg, profileID)
			}
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	configs, _ := repo.ListConfigsByProfile(int(profileID))
	if len(configs) == 0 {
		t.Error("No configs saved during concurrent access")
	}
}

func TestEdgeCase_EmptyDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "empty.db")

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}
	defer db.Close()

	repo := repository.NewFromDB(db)

	configs, err := repo.ListConfigs("")
	if err != nil {
		t.Errorf("ListConfigs failed: %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(configs))
	}

	profiles, err := repo.ListProfiles()
	if err != nil {
		t.Errorf("ListProfiles failed: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("Expected 0 profiles, got %d", len(profiles))
	}

	count, err := repo.CountConfigs()
	if err != nil {
		t.Errorf("CountConfigs failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestEdgeCase_DuplicateHandling(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "dupes.db")

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}
	defer db.Close()

	repo := repository.NewFromDB(db)
	profileID, _ := repo.GetOrCreateProfile("Test", "test", "manual")

	cfg := &common.NormalizedConfig{
		Name:     "Duplicate",
		Protocol: common.Vless,
		Server:   "example.com",
		Port:     443,
		UUID:     "12345678-1234-1234-1234-123456789012",
	}

	for i := 0; i < 100; i++ {
		repo.InsertConfigOrSkip(cfg, profileID)
	}

	configs, _ := repo.ListConfigsByProfile(int(profileID))
	if len(configs) != 1 {
		t.Errorf("Expected 1 config after 100 inserts, got %d", len(configs))
	}
}

func TestEdgeCase_InvalidPorts(t *testing.T) {
	invalidPorts := []int{-1, 0, 65536, 99999, 2147483647}

	for _, port := range invalidPorts {
		cfg := &common.RawConfig{
			Protocol: common.Vless,
			Server:   "example.com",
			Port:     port,
			UUID:     "12345678-1234-1234-1234-123456789012",
		}

		err := normalizer.Validate(cfg)
		if err == nil {
			t.Errorf("Expected error for port %d, got nil", port)
		}
	}
}

func TestEdgeCase_SpecialCharactersInServer(t *testing.T) {
	special := []string{
		"server@evil.com",
		"server:443",
		"server/path",
		"server\\path",
		"server?query",
		"server#fragment",
	}

	for _, server := range special {
		cfg := &common.RawConfig{
			Protocol: common.Vless,
			Server:   server,
			Port:     443,
			UUID:     "12345678-1234-1234-1234-123456789012",
		}

		err := normalizer.Validate(cfg)

		_ = err
	}
}

func TestEdgeCase_DatabaseCorruption(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "corrupt.db")

	db, _ := storage.Open(dbPath)
	db.Close()

	f, _ := os.OpenFile(dbPath, os.O_WRONLY, 0644)
	f.WriteString("CORRUPTED DATA")
	f.Close()

	_, err := storage.Open(dbPath)
	if err == nil {
		t.Error("Expected error opening corrupted database")
	}
}

func TestEdgeCase_VeryLargeProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large profile test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "large.db")

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}
	defer db.Close()

	repo := repository.NewFromDB(db)
	profileID, _ := repo.GetOrCreateProfile("Large", "test", "manual")

	configs := make([]*common.NormalizedConfig, 10000)
	for i := 0; i < 10000; i++ {
		configs[i] = &common.NormalizedConfig{
			Name:     fmt.Sprintf("Config-%d", i),
			Protocol: common.Vless,
			Server:   fmt.Sprintf("server%d.com", i),
			Port:     443,
			UUID:     "12345678-1234-1234-1234-123456789012",
		}
	}

	inserted, err := repo.InsertConfigBatch(configs, profileID)
	if err != nil {
		t.Fatalf("InsertConfigBatch failed: %v", err)
	}
	if inserted != 10000 {
		t.Errorf("Expected 10000 inserted, got %d", inserted)
	}

	result, err := repo.ListConfigsByProfile(int(profileID))
	if err != nil {
		t.Fatalf("ListConfigsByProfile failed: %v", err)
	}
	if len(result) != 10000 {
		t.Errorf("Expected 10000 configs, got %d", len(result))
	}
}

func TestEdgeCase_MixedValidInvalid(t *testing.T) {
	configs := []*common.RawConfig{
		{Protocol: common.Vless, Server: "valid.com", Port: 443, UUID: "12345678-1234-1234-1234-123456789012"},
		{Protocol: common.Vless, Server: "127.0.0.1", Port: 443, UUID: "12345678-1234-1234-1234-123456789012"},
		{Protocol: common.Vless, Server: "valid2.com", Port: 443, UUID: "12345678-1234-1234-1234-123456789012"},
		{Protocol: common.Vless, Server: "valid3.com", Port: 99999, UUID: "12345678-1234-1234-1234-123456789012"},
		{Protocol: common.Vless, Server: "valid4.com", Port: 443, UUID: "not-a-uuid"},
	}

	normalized, err := normalizer.Normalize(configs)
	if err != nil {
		t.Fatalf("Normalize failed: %v", err)
	}

	if len(normalized) != 2 {
		t.Errorf("Expected 2 valid configs, got %d", len(normalized))
	}
}
