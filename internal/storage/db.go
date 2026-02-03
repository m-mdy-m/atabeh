package storage

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Open(path string) (*DB, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("storage: resolve path: %w", err)
	}

	conn, err := sql.Open("sqlite", abs+"?_journal_mode=WAL&_foreign_keys=on")
	// https://stackoverflow.com/questions/31952791/setmaxopenconns-and-setmaxidleconns
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	if err != nil {
		return nil, fmt.Errorf("storage: open db: %w", err)
	}

	// Sanity ping
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("storage: ping db: %w", err)
	}

	db := &DB{inner: conn, Path: abs}

	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, err
	}

	return db, nil
}

func (db *DB) Close() error { return db.inner.Close() }
func (db *DB) Raw() *sql.DB { return db.inner }
func (db *DB) migrate() error {
	_, err := db.inner.Exec(`
		CREATE TABLE IF NOT EXISTS configs (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
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
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE UNIQUE INDEX IF NOT EXISTS idx_configs_dedup
			ON configs (protocol, server, port, uuid, password);

		CREATE INDEX IF NOT EXISTS idx_configs_protocol ON configs (protocol);
		CREATE INDEX IF NOT EXISTS idx_configs_alive    ON configs (is_alive);
	`)
	if err != nil {
		return fmt.Errorf("storage: migrate: %w", err)
	}
	return nil
}
