package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/core"
)

func (r *Repo) InsertConfig(cfg *common.NormalizedConfig, profileID int64) (int64, error) {
	extraJSON, err := core.MarshalExtra(cfg.Extra)
	if err != nil {
		return 0, err
	}

	source := fmt.Sprintf("profile:%d", profileID)

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
			cfg.Security, extraJSON, source,
		)

	return r.core.InsertQuery(q)
}

func (r *Repo) InsertConfigOrSkip(cfg *common.NormalizedConfig, profileID int64) (int64, bool, error) {
	extraJSON, err := core.MarshalExtra(cfg.Extra)
	if err != nil {
		return 0, false, err
	}

	source := fmt.Sprintf("profile:%d", profileID)

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
			cfg.Security, extraJSON, source,
		)

	id, err := r.core.InsertQuery(q)
	if err != nil {
		return 0, false, fmt.Errorf("insert config: %w", err)
	}

	if id == 0 {
		dupID, err := r.core.FindDupID(cfg)
		return dupID, false, err
	}

	return id, true, nil
}

func (r *Repo) InsertConfigBatch(configs []*common.NormalizedConfig, profileID int64) (inserted int, err error) {
	err = core.WithTx(r.core.DB, func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(`
			INSERT OR IGNORE INTO configs 
			(profile_id, name, protocol, server, port, uuid, password, method, transport, security, extra, source) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return fmt.Errorf("prepare statement: %w", err)
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

func (r *Repo) GetConfigByID(id int) (*storage.ConfigRow, error) {
	q := core.Select(
		core.ConfigColID, core.ConfigColProfileID, core.ConfigColName,
		core.ConfigColProtocol, core.ConfigColServer, core.ConfigColPort,
		core.ConfigColUUID, core.ConfigColPassword, core.ConfigColMethod,
		core.ConfigColTransport, core.ConfigColSecurity, core.ConfigColExtra,
		core.ConfigColSource, core.ConfigColLastPing, core.ConfigColIsAlive,
		core.ConfigColCreatedAt, core.ConfigColUpdatedAt,
	).
		From(core.TableConfigs).
		Where(core.ConfigColID+" = ?", id).
		Limit(1)

	sqlStr, args := q.Build()
	row := r.core.DB.QueryRow(sqlStr, args...)
	return core.GetOne(row, ScanConfig)
}

func (r *Repo) ListConfigs(protocol common.Kind) ([]*storage.ConfigRow, error) {
	q := core.Select(
		core.ConfigColID, core.ConfigColProfileID, core.ConfigColName,
		core.ConfigColProtocol, core.ConfigColServer, core.ConfigColPort,
		core.ConfigColUUID, core.ConfigColPassword, core.ConfigColMethod,
		core.ConfigColTransport, core.ConfigColSecurity, core.ConfigColExtra,
		core.ConfigColSource, core.ConfigColLastPing, core.ConfigColIsAlive,
		core.ConfigColCreatedAt, core.ConfigColUpdatedAt,
	).
		From(core.TableConfigs).
		OrderBy(core.ConfigColProfileID + ", " + core.ConfigColID)

	if protocol != "" {
		q = q.Where(core.ConfigColProtocol+" = ?", string(protocol))
	}

	sqlStr, args := q.Build()
	rows, err := r.core.DB.Query(sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("list configs: %w", err)
	}
	defer rows.Close()

	return core.List(rows, ScanConfig)
}

func (r *Repo) ListConfigsByProfile(profileID int) ([]*storage.ConfigRow, error) {
	q := core.Select(
		core.ConfigColID, core.ConfigColProfileID, core.ConfigColName,
		core.ConfigColProtocol, core.ConfigColServer, core.ConfigColPort,
		core.ConfigColUUID, core.ConfigColPassword, core.ConfigColMethod,
		core.ConfigColTransport, core.ConfigColSecurity, core.ConfigColExtra,
		core.ConfigColSource, core.ConfigColLastPing, core.ConfigColIsAlive,
		core.ConfigColCreatedAt, core.ConfigColUpdatedAt,
	).
		From(core.TableConfigs).
		Where(core.ConfigColProfileID+" = ?", profileID).
		OrderBy(core.ConfigColIsAlive + " DESC, " + core.ConfigColLastPing + " ASC")

	sqlStr, args := q.Build()
	rows, err := r.core.DB.Query(sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("list configs by profile: %w", err)
	}
	defer rows.Close()

	return core.List(rows, ScanConfig)
}

func (r *Repo) ListAliveConfigs() ([]*storage.ConfigRow, error) {
	q := core.Select(
		core.ConfigColID, core.ConfigColProfileID, core.ConfigColName,
		core.ConfigColProtocol, core.ConfigColServer, core.ConfigColPort,
		core.ConfigColUUID, core.ConfigColPassword, core.ConfigColMethod,
		core.ConfigColTransport, core.ConfigColSecurity, core.ConfigColExtra,
		core.ConfigColSource, core.ConfigColLastPing, core.ConfigColIsAlive,
		core.ConfigColCreatedAt, core.ConfigColUpdatedAt,
	).
		From(core.TableConfigs).
		Where(core.ConfigColIsAlive+" = ?", 1).
		OrderBy(core.ConfigColLastPing + " ASC")

	sqlStr, args := q.Build()
	rows, err := r.core.DB.Query(sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("list alive configs: %w", err)
	}
	defer rows.Close()

	return core.List(rows, ScanConfig)
}

func (r *Repo) CountConfigs() (int, error) {
	q := core.Select("COUNT(*)").From(core.TableConfigs)
	sqlStr, args := q.Build()

	var n int
	err := r.core.DB.QueryRow(sqlStr, args...).Scan(&n)
	return n, err
}

func (r *Repo) CountConfigsByProfile(profileID int) (int, error) {
	q := core.Select("COUNT(*)").
		From(core.TableConfigs).
		Where(core.ConfigColProfileID+" = ?", profileID)

	sqlStr, args := q.Build()
	var n int
	err := r.core.DB.QueryRow(sqlStr, args...).Scan(&n)
	return n, err
}

func (r *Repo) UpdateConfigPingResult(id int, result *common.PingResult) error {
	q := core.Update(core.TableConfigs).
		Set(core.ConfigColLastPing, result.AvgMs).
		Set(core.ConfigColIsAlive, core.BoolToInt(result.Reachable)).
		Set(core.ConfigColUpdatedAt, time.Now()).
		Where(core.ConfigColID+" = ?", id)

	_, err := r.core.ExecQuery(q)
	if err != nil {
		return fmt.Errorf("update ping result: %w", err)
	}
	return nil
}

func (r *Repo) UpdateConfigPingBatch(results map[int]*common.PingResult) error {
	return core.WithTx(r.core.DB, func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(
			"UPDATE configs SET last_ping=?, is_alive=?, updated_at=? WHERE id=?")
		if err != nil {
			return fmt.Errorf("prepare statement: %w", err)
		}
		defer stmt.Close()

		now := time.Now()
		for id, result := range results {
			_, err := stmt.Exec(
				result.AvgMs,
				core.BoolToInt(result.Reachable),
				now,
				id,
			)
			if err != nil {
				return fmt.Errorf("update config %d: %w", id, err)
			}
		}
		return nil
	})
}

func (r *Repo) DeleteConfigByID(id int) error {
	q := core.DeleteFrom(core.TableConfigs).
		Where(core.ConfigColID+" = ?", id)

	affected, err := r.core.ExecQuery(q)
	if err != nil {
		return fmt.Errorf("delete config: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("config id %d not found", id)
	}
	return nil
}

func (r *Repo) DeleteConfigsByProfile(profileID int) (int, error) {
	q := core.DeleteFrom(core.TableConfigs).
		Where(core.ConfigColProfileID+" = ?", profileID)

	affected, err := r.core.ExecQuery(q)
	if err != nil {
		return 0, fmt.Errorf("delete configs by profile: %w", err)
	}
	return int(affected), nil
}

func (r *Repo) ClearAllConfigs() error {
	q := core.DeleteFrom(core.TableConfigs)
	_, err := r.core.ExecQuery(q)
	return err
}
