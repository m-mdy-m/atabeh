package repository

import (
	"database/sql"
	"fmt"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/tester"
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/core"
)

func (r *Repo) InsertConfig(cfg *common.NormalizedConfig, profileID int64) (int64, error) {
	extraJSON, err := core.MarshalExtra(cfg.Extra)
	if err != nil {
		return 0, err
	}

	q := core.InsertInto(core.TableConfigs).
		Columns(
			core.ConfigColProfileID, core.ConfigColName, core.ConfigColProtocol,
			core.ConfigColServer, core.ConfigColPort, core.ConfigColUUID,
			core.ConfigColPassword, core.ConfigColMethod, core.ConfigColTransport,
			core.ConfigColSecurity, core.ConfigColExtra, core.ConfigColSource,
		).
		Values(
			profileID, cfg.Name, string(cfg.Protocol),
			cfg.Server, cfg.Port, cfg.UUID,
			cfg.Password, cfg.Method, string(cfg.Transport),
			cfg.Security, extraJSON, fmt.Sprintf("profile:%d", profileID),
		)

	return r.core.InsertQuery(q)
}

func (r *Repo) InsertConfigBatch(configs []*common.NormalizedConfig, profileID int64) (int, error) {
	inserted := 0

	err := core.WithTx(r.core.DB, func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(`
			INSERT OR IGNORE INTO configs 
			(profile_id, name, protocol, server, port, uuid, password, method, transport, security, extra, source) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return fmt.Errorf("prepare: %w", err)
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
		return nil
	})

	return inserted, err
}

func (r *Repo) InsertConfigOrSkip(cfg *common.NormalizedConfig, profileID int64) (int64, bool, error) {
	extraJSON, err := core.MarshalExtra(cfg.Extra)
	if err != nil {
		return 0, false, err
	}

	q := core.InsertInto(core.TableConfigs).
		OrIgnore().
		Columns(
			core.ConfigColProfileID, core.ConfigColName, core.ConfigColProtocol,
			core.ConfigColServer, core.ConfigColPort, core.ConfigColUUID,
			core.ConfigColPassword, core.ConfigColMethod, core.ConfigColTransport,
			core.ConfigColSecurity, core.ConfigColExtra, core.ConfigColSource,
		).
		Values(
			profileID, cfg.Name, string(cfg.Protocol),
			cfg.Server, cfg.Port, cfg.UUID,
			cfg.Password, cfg.Method, string(cfg.Transport),
			cfg.Security, extraJSON, fmt.Sprintf("profile:%d", profileID),
		)

	id, err := r.core.InsertQuery(q)
	if err != nil {
		return 0, false, fmt.Errorf("insert: %w", err)
	}

	if id == 0 {
		dupID, err := r.core.FindDupID(cfg)
		return dupID, false, err
	}

	return id, true, nil
}

func (r *Repo) GetConfigByID(id int) (*storage.ConfigRow, error) {
	q := core.Select("*").From(core.TableConfigs).Where("id = ?", id).Limit(1)

	sqlStr, args := q.Build()
	row := r.core.DB.QueryRow(sqlStr, args...)
	return core.GetOne(row, scanConfig)
}

func (r *Repo) ListConfigs(protocol common.Kind) ([]*storage.ConfigRow, error) {
	q := core.Select("*").From(core.TableConfigs).OrderBy("profile_id, id")

	if protocol != "" {
		q = q.Where("protocol = ?", string(protocol))
	}

	sqlStr, args := q.Build()
	rows, err := r.core.DB.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return core.List(rows, scanConfig)
}

func (r *Repo) ListConfigsByProfile(profileID int) ([]*storage.ConfigRow, error) {
	q := core.Select("*").
		From(core.TableConfigs).
		Where("profile_id = ?", profileID).
		OrderBy("is_alive DESC, last_ping ASC")

	sqlStr, args := q.Build()
	rows, err := r.core.DB.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return core.List(rows, scanConfig)
}

func (r *Repo) ListAliveConfigs() ([]*storage.ConfigRow, error) {
	q := core.Select("*").
		From(core.TableConfigs).
		Where("is_alive = ?", 1).
		OrderBy("last_ping ASC")

	sqlStr, args := q.Build()
	rows, err := r.core.DB.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return core.List(rows, scanConfig)
}

func (r *Repo) CountConfigs() (int, error) {
	var n int
	err := r.core.DB.QueryRow("SELECT COUNT(*) FROM configs").Scan(&n)
	return n, err
}

func (r *Repo) CountConfigsByProfile(profileID int) (int, error) {
	var n int
	err := r.core.DB.QueryRow("SELECT COUNT(*) FROM configs WHERE profile_id = ?", profileID).Scan(&n)
	return n, err
}

func (r *Repo) UpdateConfigPingResult(id int, result *tester.Result) error {
	_, err := r.core.DB.Exec(
		"UPDATE configs SET last_ping=?, is_alive=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		result.AvgMs, core.BoolToInt(result.Reachable), id)
	return err
}

func (r *Repo) UpdateConfigPingBatch(results map[int]*tester.Result) error {
	return core.WithTx(r.core.DB, func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(
			"UPDATE configs SET last_ping=?, is_alive=?, updated_at=CURRENT_TIMESTAMP WHERE id=?")
		if err != nil {
			return err
		}
		defer stmt.Close()

		for id, result := range results {
			_, err := stmt.Exec(result.AvgMs, core.BoolToInt(result.Reachable), id)
			if err != nil {
				return fmt.Errorf("update config %d: %w", id, err)
			}
		}
		return nil
	})
}

func (r *Repo) DeleteConfigByID(id int) error {
	res, err := r.core.DB.Exec("DELETE FROM configs WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("config %d not found", id)
	}
	return nil
}

func (r *Repo) ClearAllConfigs() error {
	_, err := r.core.DB.Exec("DELETE FROM configs")
	return err
}
