package core

import (
	"database/sql"
)

type Repo struct {
	DB   *sql.DB
	Path string
}
type Scanner interface {
	Scan(dest ...any) error
}

func New(db *sql.DB) *Repo {
	return &Repo{DB: db}
}
