package repository

import (
	"database/sql"
	"fmt"

	"github.com/m-mdy-m/atabeh/storage"
)

func (r *Repo) CreateProfile(name, source, profileType string) (int64, error) {
	res, err := r.core.DB.Exec(
		"INSERT INTO profiles (name, source, type, last_synced_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)",
		name, source, profileType)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repo) GetOrCreateProfile(name, source, profileType string) (int64, error) {
	var id int64
	err := r.core.DB.QueryRow("SELECT id FROM profiles WHERE source = ?", source).Scan(&id)

	if err == sql.ErrNoRows {
		return r.CreateProfile(name, source, profileType)
	}
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repo) GetProfile(id int) (*storage.ProfileRow, error) {
	var p storage.ProfileRow
	var lastSynced sql.NullString

	err := r.core.DB.QueryRow(`
		SELECT id, name, source, type, config_count, alive_count, 
		       last_synced_at, created_at, updated_at
		FROM profiles WHERE id = ?`, id).Scan(
		&p.ID, &p.Name, &p.Source, &p.Type,
		&p.ConfigCount, &p.AliveCount,
		&lastSynced, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, err
	}

	if lastSynced.Valid {
		p.LastSyncedAt = lastSynced.String
	}

	return &p, nil
}

func (r *Repo) ListProfiles() ([]*storage.ProfileRow, error) {
	rows, err := r.core.DB.Query(`
		SELECT id, name, source, type, config_count, alive_count,
		       last_synced_at, created_at, updated_at
		FROM profiles ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*storage.ProfileRow
	for rows.Next() {
		var p storage.ProfileRow
		var lastSynced sql.NullString

		err := rows.Scan(
			&p.ID, &p.Name, &p.Source, &p.Type,
			&p.ConfigCount, &p.AliveCount,
			&lastSynced, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if lastSynced.Valid {
			p.LastSyncedAt = lastSynced.String
		}
		profiles = append(profiles, &p)
	}
	return profiles, rows.Err()
}

func (r *Repo) UpdateProfileSyncTime(id int64) error {
	_, err := r.core.DB.Exec(
		"UPDATE profiles SET last_synced_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id=?", id)
	return err
}

func (r *Repo) DeleteProfile(id int) error {
	res, err := r.core.DB.Exec("DELETE FROM profiles WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("profile %d not found", id)
	}
	return nil
}
