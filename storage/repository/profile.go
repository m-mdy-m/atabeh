package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/core"
)

func (r *Repo) CreateProfile(name, source, profileType string) (int64, error) {
	q := core.InsertInto(core.TableProfiles).
		Columns(
			core.ProfileColName,
			core.ProfileColSource,
			core.ProfileColType,
			core.ProfileColLastSynced,
		).
		Values(name, source, profileType, time.Now())

	id, err := r.core.InsertQuery(q)
	if err != nil {
		return 0, fmt.Errorf("create profile: %w", err)
	}
	return id, nil
}

func (r *Repo) GetOrCreateProfile(name, source, profileType string) (int64, error) {

	q := core.Select(core.ProfileColID).
		From(core.TableProfiles).
		Where(core.ProfileColSource+" = ?", source).
		Limit(1)

	sqlStr, args := q.Build()

	var id int64
	err := r.core.DB.QueryRow(sqlStr, args...).Scan(&id)

	if err == sql.ErrNoRows {

		return r.CreateProfile(name, source, profileType)
	}
	if err != nil {
		return 0, fmt.Errorf("get profile: %w", err)
	}

	return id, nil
}

func (r *Repo) ListProfiles() ([]*storage.ProfileRow, error) {
	q := core.Select(
		core.ProfileColID,
		core.ProfileColName,
		core.ProfileColSource,
		core.ProfileColType,
		core.ProfileColConfigCount,
		core.ProfileColAliveCount,
		core.ProfileColLastSynced,
		core.ProfileColCreatedAt,
		core.ProfileColUpdatedAt,
	).
		From(core.TableProfiles).
		OrderBy(core.ProfileColUpdatedAt + " DESC")

	sqlStr, args := q.Build()
	rows, err := r.core.DB.Query(sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	defer rows.Close()

	var profiles []*storage.ProfileRow
	for rows.Next() {
		var p storage.ProfileRow
		var lastSynced sql.NullString

		err := rows.Scan(
			&p.ID, &p.Name, &p.Source, &p.Type,
			&p.ConfigCount, &p.AliveCount,
			&lastSynced, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan profile: %w", err)
		}

		if lastSynced.Valid {
			p.LastSyncedAt = lastSynced.String
		}
		profiles = append(profiles, &p)
	}
	return profiles, rows.Err()
}

func (r *Repo) GetProfile(id int) (*storage.ProfileRow, error) {
	q := core.Select(
		core.ProfileColID,
		core.ProfileColName,
		core.ProfileColSource,
		core.ProfileColType,
		core.ProfileColConfigCount,
		core.ProfileColAliveCount,
		core.ProfileColLastSynced,
		core.ProfileColCreatedAt,
		core.ProfileColUpdatedAt,
	).
		From(core.TableProfiles).
		Where(core.ProfileColID+" = ?", id).
		Limit(1)

	sqlStr, args := q.Build()

	var p storage.ProfileRow
	var lastSynced sql.NullString

	err := r.core.DB.QueryRow(sqlStr, args...).Scan(
		&p.ID, &p.Name, &p.Source, &p.Type,
		&p.ConfigCount, &p.AliveCount,
		&lastSynced, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}

	if lastSynced.Valid {
		p.LastSyncedAt = lastSynced.String
	}

	return &p, nil
}

func (r *Repo) UpdateProfileSyncTime(id int64) error {
	q := core.Update(core.TableProfiles).
		Set(core.ProfileColLastSynced, time.Now()).
		Set(core.ProfileColUpdatedAt, time.Now()).
		Where(core.ProfileColID+" = ?", id)

	_, err := r.core.ExecQuery(q)
	return err
}

func (r *Repo) DeleteProfile(id int) error {
	q := core.DeleteFrom(core.TableProfiles).
		Where(core.ProfileColID+" = ?", id)

	_, err := r.core.ExecQuery(q)
	return err
}
