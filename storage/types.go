package storage

import "github.com/m-mdy-m/atabeh/internal/common"

type ConfigRow struct {
	ID        int         `json:"id"`
	ProfileID int         `json:"profile_id"`
	Name      string      `json:"name"`
	Protocol  common.Kind `json:"protocol"`
	Server    string      `json:"server"`
	Port      int         `json:"port"`
	UUID      string      `json:"uuid"`
	Password  string      `json:"password"`
	Method    string      `json:"method"`
	Transport common.Kind `json:"transport"`
	Security  string      `json:"security"`
	Extra     string      `json:"extra"`  // JSON-serialised map
	Source    string      `json:"source"` // subscription URL or "manual"
	LastPing  int64       `json:"last_ping_ms"`
	IsAlive   bool        `json:"is_alive"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
}

type ProfileRow struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Source       string `json:"source"`
	Type         string `json:"type"` // "subscription", "manual", "mixed"
	ConfigCount  int    `json:"config_count"`
	AliveCount   int    `json:"alive_count"`
	LastSyncedAt string `json:"last_synced_at"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
