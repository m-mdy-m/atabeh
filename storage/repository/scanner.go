package repository

import (
	"database/sql"

	"github.com/m-mdy-m/atabeh/internal/logger"
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/core"
)

func ScanConfig(s core.Scanner) (*storage.ConfigRow, error) {
	var c storage.ConfigRow
	var extraRaw sql.NullString
	var isAlive sql.NullInt64
	var lastPing sql.NullInt64

	err := s.Scan(
		&c.ID, &c.ProfileID, &c.Name, &c.Protocol, &c.Server, &c.Port,
		&c.UUID, &c.Password, &c.Method, &c.Transport, &c.Security,
		&extraRaw, &c.Source, &lastPing, &isAlive, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, logger.Errorf("DATABASE", "scan configs: %w", err)
	}

	c.LastPing = 0
	if lastPing.Valid {
		c.LastPing = lastPing.Int64
	}

	c.IsAlive = false
	if isAlive.Valid && isAlive.Int64 == 1 {
		c.IsAlive = true
	}

	if extraRaw.Valid {
		c.Extra = extraRaw.String
	} else {
		c.Extra = "{}"
	}

	return &c, nil
}

func ScanConfigs(rows *sql.Rows) ([]*storage.ConfigRow, error) {
	var out []*storage.ConfigRow
	for rows.Next() {
		cfg, err := ScanConfig(rows)
		if err != nil {
			return nil, logger.Errorf("DATABASE", "scan configs: %w", err)
		}
		out = append(out, cfg)
	}
	if err := rows.Err(); err != nil {
		return nil, logger.Errorf("DATABASE", "scan configs rows error: %w", err)
	}
	return out, nil
}
