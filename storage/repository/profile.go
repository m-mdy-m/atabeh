package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/core"
)

func (r *Repo) CreateProfile(name, source, profileType string) (int64, error) {
	q := core.InsertInto(core.TableProfiles).Columns("name", "source", "type", "last_synced_at").Values(name, source, profileType, time.Now())
	id, err := r.core.InsertQuery(q)

	if err != nil {
		return 0, fmt.Errorf("storage: create profile: %w", err)
	}
	return id, err
}

func (r *Repo) GetOrCreateProfile(name, source, profileType string) (int64, error) {

	var id int64
	err := r.db.Raw().QueryRow(
		"SELECT id FROM profiles WHERE source = ?", source).Scan(&id)

	if err == sql.ErrNoRows {
		return r.CreateProfile(name, source, profileType)
	}
	if err != nil {
		return 0, fmt.Errorf("storage: get profile: %w", err)
	}
	return id, nil
}

func (r *Repo) ListProfiles() ([]*storage.ProfileRow, error) {
	rows, err := r.db.Raw().Query(`
		SELECT id, name, source, type, config_count, alive_count, 
		       last_synced_at, created_at, updated_at 
		FROM profiles 
		ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("storage: list profiles: %w", err)
	}
	defer rows.Close()

	var profiles []*storage.ProfileRow
	for rows.Next() {
		var p storage.ProfileRow
		var lastSynced sql.NullString
		err := rows.Scan(&p.ID, &p.Name, &p.Source, &p.Type, &p.ConfigCount, &p.AliveCount,
			&lastSynced, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("storage: scan profile: %w", err)
		}
		if lastSynced.Valid {
			p.LastSyncedAt = lastSynced.String
		}
		profiles = append(profiles, &p)
	}
	return profiles, rows.Err()
}

func (r *Repo) GetProfile(id int) (*storage.ProfileRow, error) {
	var p storage.ProfileRow
	var lastSynced sql.NullString
	err := r.db.Raw().QueryRow(`
		SELECT id, name, source, type, config_count, alive_count, 
		       last_synced_at, created_at, updated_at 
		FROM profiles WHERE id = ?`, id).Scan(
		&p.ID, &p.Name, &p.Source, &p.Type, &p.ConfigCount, &p.AliveCount,
		&lastSynced, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("storage: get profile: %w", err)
	}
	if lastSynced.Valid {
		p.LastSyncedAt = lastSynced.String
	}
	return &p, nil
}

func (r *Repo) UpdateProfileSyncTime(id int64) error {
	_, err := r.db.Raw().Exec(
		"UPDATE profiles SET last_synced_at = ?, updated_at = ? WHERE id = ?",
		time.Now(), time.Now(), id)
	return err
}

func (r *Repo) DeleteProfile(id int) error {
	_, err := r.db.Raw().Exec("DELETE FROM profiles WHERE id = ?", id)
	return err
}
