package storage

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/m-mdy-m/atabeh/storage/core"
	_ "modernc.org/sqlite"
)

type RepoDatabase struct {
	*core.Repo
}

func Open(path string) (*RepoDatabase, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("storage: resolve path: %w", err)
	}

	conn, err := sql.Open("sqlite", abs+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("storage: open db: %w", err)
	}
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("storage: ping db: %w", err)
	}

	db := &RepoDatabase{
		Repo: &core.Repo{
			DB:   conn,
			Path: abs,
		},
	}

	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, err
	}

	return db, nil
}

func (db *RepoDatabase) Close() error { return db.DB.Close() }
func (db *RepoDatabase) Raw() *sql.DB { return db.DB }

func (db *RepoDatabase) migrate() error {
	_, err := db.DB.Exec(`
		-- Profiles table
		CREATE TABLE IF NOT EXISTS profiles (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			name            TEXT    NOT NULL,
			source          TEXT    NOT NULL UNIQUE,
			type            TEXT    NOT NULL DEFAULT 'mixed',
			config_count    INTEGER NOT NULL DEFAULT 0,
			alive_count     INTEGER NOT NULL DEFAULT 0,
			last_synced_at  DATETIME,
			created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- Configs table with profile_id
		CREATE TABLE IF NOT EXISTS configs (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_id INTEGER NOT NULL,
			name       TEXT    NOT NULL DEFAULT '',
			protocol   TEXT    NOT NULL,
			server     TEXT    NOT NULL,
			port       INTEGER NOT NULL,
			uuid       TEXT    NOT NULL DEFAULT '',
			password   TEXT    NOT NULL DEFAULT '',
			method     TEXT    NOT NULL DEFAULT '',
			transport  TEXT    NOT NULL DEFAULT 'tcp',
			security   TEXT    NOT NULL DEFAULT 'none',
			extra      TEXT    NOT NULL DEFAULT '{}',
			source     TEXT    NOT NULL DEFAULT 'manual',
			last_ping  INTEGER NOT NULL DEFAULT -1,
			is_alive   INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
		);

		-- Indexes for configs
		CREATE UNIQUE INDEX IF NOT EXISTS idx_configs_dedup
			ON configs (protocol, server, port, uuid, password);

		CREATE INDEX IF NOT EXISTS idx_configs_profile   ON configs (profile_id);
		CREATE INDEX IF NOT EXISTS idx_configs_protocol  ON configs (protocol);
		CREATE INDEX IF NOT EXISTS idx_configs_alive     ON configs (is_alive);

		-- Indexes for profiles
		CREATE INDEX IF NOT EXISTS idx_profiles_source ON profiles (source);

		-- Subscriptions table
		CREATE TABLE IF NOT EXISTS subscriptions (
			url TEXT PRIMARY KEY
		);

		-- Triggers to update profile counts
		CREATE TRIGGER IF NOT EXISTS update_profile_counts_insert
		AFTER INSERT ON configs
		BEGIN
			UPDATE profiles 
			SET config_count = config_count + 1,
			    alive_count = alive_count + NEW.is_alive,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = NEW.profile_id;
		END;

		CREATE TRIGGER IF NOT EXISTS update_profile_counts_update
		AFTER UPDATE ON configs
		BEGIN
			UPDATE profiles 
			SET alive_count = (
				SELECT COUNT(*) FROM configs 
				WHERE profile_id = NEW.profile_id AND is_alive = 1
			),
			updated_at = CURRENT_TIMESTAMP
			WHERE id = NEW.profile_id;
		END;

		CREATE TRIGGER IF NOT EXISTS update_profile_counts_delete
		AFTER DELETE ON configs
		BEGIN
			UPDATE profiles 
			SET config_count = config_count - 1,
			    alive_count = alive_count - OLD.is_alive,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = OLD.profile_id;
		END;
	`)
	if err != nil {
		return fmt.Errorf("storage: migrate: %w", err)
	}
	return nil
}
