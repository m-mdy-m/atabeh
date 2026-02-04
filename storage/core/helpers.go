package core

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/m-mdy-m/atabeh/internal/common"
)

func (r *Repo) FindDupID(cfg *common.NormalizedConfig) (int64, error) {
	var id int64

	query, args := Select("id").
		From("configs").
		Where("protocol = ? AND server = ? AND port = ? AND uuid = ? AND password = ?",
			string(cfg.Protocol), cfg.Server, cfg.Port, cfg.UUID, cfg.Password).
		Limit(1).
		Build()

	row := r.DB.QueryRow(query, args...)
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("storage: find dup: %w", err)
	}
	return id, nil
}
func MarshalExtra(m map[string]string) (string, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("storage: marshal extra: %w", err)
	}
	return string(b), nil
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
