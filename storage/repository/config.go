package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/core"
)

func (r *Repo) Insert(cfg *common.NormalizedConfig, profileID int64) (int64, error) {
	extraJSON, err := core.MarshalExtra(cfg.Extra)
	if err != nil {
		return 0, err
	}

	source := fmt.Sprintf("profile:%d", profileID)
	res, err := r.core.DB.Exec(
		`INSERT INTO configs 
		(profile_id, name, protocol, server, port, uuid, password, method, transport, security, extra, source) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		profileID, cfg.Name, string(cfg.Protocol), cfg.Server, cfg.Port,
		cfg.UUID, cfg.Password, cfg.Method,
		string(cfg.Transport), cfg.Security, extraJSON, source)
	if err != nil {
		return 0, fmt.Errorf("storage: insert config: %w", err)
	}
	return res.LastInsertId()
}

func (r *Repo) InsertOrSkip(cfg *common.NormalizedConfig, profileID int64) (int64, bool, error) {
	extraJSON, err := core.MarshalExtra(cfg.Extra)
	if err != nil {
		return 0, false, err
	}

	source := fmt.Sprintf("profile:%d", profileID)
	res, err := r.core.Raw().Exec(
		`INSERT OR IGNORE INTO configs 
		(profile_id, name, protocol, server, port, uuid, password, method, transport, security, extra, source) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		profileID, cfg.Name, string(cfg.Protocol), cfg.Server, cfg.Port,
		cfg.UUID, cfg.Password, cfg.Method,
		string(cfg.Transport), cfg.Security, extraJSON, source)
	if err != nil {
		return 0, false, fmt.Errorf("storage: insert-or-skip: %w", err)
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		id, err := r.FindDupID(cfg)
		return id, false, err
	}
	id, _ := res.LastInsertId()
	return id, true, nil
}

func (r *Repo) InsertBatch(configs []*common.NormalizedConfig, profileID int64) (inserted int, err error) {
	tx, err := r.core.Raw().Begin()
	if err != nil {
		return 0, fmt.Errorf("storage: begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO configs 
		(profile_id, name, protocol, server, port, uuid, password, method, transport, security, extra, source) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("storage: prepare statement: %w", err)
	}
	defer stmt.Close()

	source := fmt.Sprintf("profile:%d", profileID)
	for _, cfg := range configs {
		extraJSON, err := core.MarshalExtra(cfg.Extra)
		if err != nil {
			continue
		}

		res, err := stmt.Exec(
			profileID, cfg.Name, string(cfg.Protocol), cfg.Server, cfg.Port,
			cfg.UUID, cfg.Password, cfg.Method,
			string(cfg.Transport), cfg.Security, extraJSON, source)
		if err != nil {
			continue
		}

		affected, _ := res.RowsAffected()
		if affected > 0 {
			inserted++
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("storage: commit transaction: %w", err)
	}

	return inserted, nil
}

func (r *Repo) GetByID(id int) (*storage.ConfigRow, error) {
	row := r.core.Raw().QueryRow(`
		SELECT id, profile_id, name, protocol, server, port, uuid, password, method, 
		       transport, security, extra, source, last_ping, is_alive, created_at, updated_at 
		FROM configs WHERE id=?`, id)
	return scanConfig(row)
}

func (r *Repo) List(protocol common.Kind) ([]*storage.ConfigRow, error) {
	var rows *sql.Rows
	var err error
	if protocol == "" {
		rows, err = r.core.Raw().Query(`
			SELECT id, profile_id, name, protocol, server, port, uuid, password, method, 
			       transport, security, extra, source, last_ping, is_alive, created_at, updated_at 
			FROM configs ORDER BY profile_id, id`)
	} else {
		rows, err = r.core.Raw().Query(`
			SELECT id, profile_id, name, protocol, server, port, uuid, password, method, 
			       transport, security, extra, source, last_ping, is_alive, created_at, updated_at 
			FROM configs WHERE protocol=? ORDER BY profile_id, id`, string(protocol))
	}
	if err != nil {
		return nil, fmt.Errorf("storage: list: %w", err)
	}
	defer rows.Close()
	return scanConfigs(rows)
}

func (r *Repo) ListByProfile(profileID int) ([]*storage.ConfigRow, error) {
	rows, err := r.core.Raw().Query(`
		SELECT id, profile_id, name, protocol, server, port, uuid, password, method, 
		       transport, security, extra, source, last_ping, is_alive, created_at, updated_at 
		FROM configs WHERE profile_id = ? ORDER BY is_alive DESC, last_ping ASC`, profileID)
	if err != nil {
		return nil, fmt.Errorf("storage: list by profile: %w", err)
	}
	defer rows.Close()
	return scanConfigs(rows)
}

func (r *Repo) ListAlive() ([]*storage.ConfigRow, error) {
	rows, err := r.core.Raw().Query(`
		SELECT id, profile_id, name, protocol, server, port, uuid, password, method, 
		       transport, security, extra, source, last_ping, is_alive, created_at, updated_at 
		FROM configs WHERE is_alive=1 ORDER BY last_ping ASC`)
	if err != nil {
		return nil, fmt.Errorf("storage: list alive: %w", err)
	}
	defer rows.Close()
	return scanConfigs(rows)
}

func (r *Repo) Count() (int, error) {
	var n int
	err := r.core.Raw().QueryRow("SELECT COUNT(*) FROM configs").Scan(&n)
	return n, err
}

func (r *Repo) CountByProfile(profileID int) (int, error) {
	var n int
	err := r.core.Raw().QueryRow("SELECT COUNT(*) FROM configs WHERE profile_id = ?", profileID).Scan(&n)
	return n, err
}

func (r *Repo) UpdatePingResult(id int, result *common.PingResult) error {
	_, err := r.core.Raw().Exec(
		"UPDATE configs SET last_ping=?, is_alive=?, updated_at=? WHERE id=?",
		result.AvgMs, boolToInt(result.Reachable), time.Now(), id)
	if err != nil {
		return fmt.Errorf("storage: update ping: %w", err)
	}
	return nil
}

func (r *Repo) UpdatePingBatch(results map[int]*common.PingResult) error {
	tx, err := r.core.Raw().Begin()
	if err != nil {
		return fmt.Errorf("storage: begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE configs SET last_ping=?, is_alive=?, updated_at=? WHERE id=?")
	if err != nil {
		return fmt.Errorf("storage: prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for id, result := range results {
		_, err := stmt.Exec(result.AvgMs, boolToInt(result.Reachable), now, id)
		if err != nil {
			return fmt.Errorf("storage: update config %d: %w", id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("storage: commit transaction: %w", err)
	}

	return nil
}

func (r *Repo) DeleteByID(id int) error {
	res, err := r.core.Raw().Exec("DELETE FROM configs WHERE id=?", id)
	if err != nil {
		return fmt.Errorf("storage: delete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("storage: id %d not found", id)
	}
	return nil
}

func (r *Repo) DeleteByProfile(profileID int) (int, error) {
	res, err := r.core.Raw().Exec("DELETE FROM configs WHERE profile_id=?", profileID)
	if err != nil {
		return 0, fmt.Errorf("storage: delete by profile: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (r *Repo) Clear() error {
	_, err := r.core.Raw().Exec("DELETE FROM configs")
	return err
}
