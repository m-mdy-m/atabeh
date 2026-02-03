package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
)

func NewConfigRepo(db *DB) *ConfigRepo {
	return &ConfigRepo{db: db}
}

func (r *ConfigRepo) Insert(cfg *common.NormalizedConfig, source string) (int64, error) {
	extraJSON, err := marshalExtra(cfg.Extra)
	if err != nil {
		return 0, err
	}
	res, err := r.db.Raw().Exec(
		"INSERT INTO configs (name,protocol,server,port,uuid,password,method,transport,security,extra,source) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
		cfg.Name, string(cfg.Protocol), cfg.Server, cfg.Port,
		cfg.UUID, cfg.Password, cfg.Method,
		string(cfg.Transport), cfg.Security, extraJSON, source,
	)
	if err != nil {
		return 0, fmt.Errorf("storage: insert: %w", err)
	}
	return res.LastInsertId()
}

func (r *ConfigRepo) InsertOrSkip(cfg *common.NormalizedConfig, source string) (int64, bool, error) {
	extraJSON, err := marshalExtra(cfg.Extra)
	if err != nil {
		return 0, false, err
	}
	res, err := r.db.Raw().Exec(
		"INSERT OR IGNORE INTO configs (name,protocol,server,port,uuid,password,method,transport,security,extra,source) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
		cfg.Name, string(cfg.Protocol), cfg.Server, cfg.Port,
		cfg.UUID, cfg.Password, cfg.Method,
		string(cfg.Transport), cfg.Security, extraJSON, source,
	)
	if err != nil {
		return 0, false, fmt.Errorf("storage: insert-or-skip: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		id, err := r.findDupID(cfg)
		return id, false, err
	}
	id, _ := res.LastInsertId()
	return id, true, nil
}

func (r *ConfigRepo) GetByID(id int) (*common.StoredConfig, error) {
	row := r.db.Raw().QueryRow(
		"SELECT id,name,protocol,server,port,uuid,password,method,transport,security,extra,source,last_ping,is_alive,created_at,updated_at FROM configs WHERE id=?", id)
	return scanConfig(row)
}

func (r *ConfigRepo) List(protocol common.Kind) ([]*common.StoredConfig, error) {
	var rows *sql.Rows
	var err error
	if protocol == "" {
		rows, err = r.db.Raw().Query(
			"SELECT id,name,protocol,server,port,uuid,password,method,transport,security,extra,source,last_ping,is_alive,created_at,updated_at FROM configs ORDER BY id")
	} else {
		rows, err = r.db.Raw().Query(
			"SELECT id,name,protocol,server,port,uuid,password,method,transport,security,extra,source,last_ping,is_alive,created_at,updated_at FROM configs WHERE protocol=? ORDER BY id", string(protocol))
	}
	if err != nil {
		return nil, fmt.Errorf("storage: list: %w", err)
	}
	defer rows.Close()
	return scanConfigs(rows)
}

func (r *ConfigRepo) ListAlive() ([]*common.StoredConfig, error) {
	rows, err := r.db.Raw().Query(
		"SELECT id,name,protocol,server,port,uuid,password,method,transport,security,extra,source,last_ping,is_alive,created_at,updated_at FROM configs WHERE is_alive=1 ORDER BY last_ping ASC")
	if err != nil {
		return nil, fmt.Errorf("storage: list alive: %w", err)
	}
	defer rows.Close()
	return scanConfigs(rows)
}

func (r *ConfigRepo) Count() (int, error) {
	var n int
	err := r.db.Raw().QueryRow("SELECT COUNT(*) FROM configs").Scan(&n)
	return n, err
}

func (r *ConfigRepo) UpdatePingResult(id int, result *common.PingResult) error {
	_, err := r.db.Raw().Exec(
		"UPDATE configs SET last_ping=?,is_alive=?,updated_at=? WHERE id=?",
		result.AvgMs, boolToInt(result.Reachable), time.Now(), id)
	if err != nil {
		return fmt.Errorf("storage: update ping: %w", err)
	}
	return nil
}

func (r *ConfigRepo) DeleteByID(id int) error {
	res, err := r.db.Raw().Exec("DELETE FROM configs WHERE id=?", id)
	if err != nil {
		return fmt.Errorf("storage: delete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("storage: id %d not found", id)
	}
	return nil
}

func (r *ConfigRepo) DeleteBySource(source string) (int, error) {
	res, err := r.db.Raw().Exec("DELETE FROM configs WHERE source=?", source)
	if err != nil {
		return 0, fmt.Errorf("storage: delete by source: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (r *ConfigRepo) Clear() error {
	_, err := r.db.Raw().Exec("DELETE FROM configs")
	return err
}

func (r *ConfigRepo) AddSubscription(url string) error {
	_, err := r.db.Raw().Exec("INSERT OR IGNORE INTO subscriptions (url) VALUES (?)", url)
	return err
}

func (r *ConfigRepo) ListSubscriptions() ([]string, error) {
	rows, err := r.db.Raw().Query("SELECT url FROM subscriptions ORDER BY url")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *ConfigRepo) RemoveSubscription(url string) error {
	_, err := r.db.Raw().Exec("DELETE FROM subscriptions WHERE url=?", url)
	return err
}

func (r *ConfigRepo) ClearSubscriptions() error {
	_, err := r.db.Raw().Exec("DELETE FROM subscriptions")
	return err
}

// --- internal ---

func (r *ConfigRepo) findDupID(cfg *common.NormalizedConfig) (int64, error) {
	var id int64
	err := r.db.Raw().QueryRow(
		"SELECT id FROM configs WHERE protocol=? AND server=? AND port=? AND uuid=? AND password=? LIMIT 1",
		string(cfg.Protocol), cfg.Server, cfg.Port, cfg.UUID, cfg.Password).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("storage: find dup: %w", err)
	}
	return id, nil
}

type scannable interface{ Scan(dest ...any) error }

func scanConfig(s scannable) (*common.StoredConfig, error) {
	var c common.StoredConfig
	var extraRaw string
	var isAlive int
	var lastPing int64
	err := s.Scan(&c.ID, &c.Name, &c.Protocol, &c.Server, &c.Port,
		&c.UUID, &c.Password, &c.Method, &c.Transport, &c.Security,
		&extraRaw, &c.Source, &lastPing, &isAlive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	c.LastPing = lastPing
	c.IsAlive = isAlive == 1
	c.Extra = extraRaw
	return &c, nil
}

func scanConfigs(rows *sql.Rows) ([]*common.StoredConfig, error) {
	var out []*common.StoredConfig
	for rows.Next() {
		c, err := scanConfig(rows)
		if err != nil {
			return nil, fmt.Errorf("storage: scan: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func marshalExtra(m map[string]string) (string, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("storage: marshal extra: %w", err)
	}
	return string(b), nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
