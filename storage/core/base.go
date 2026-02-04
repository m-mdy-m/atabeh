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
	return &Repo{
		DB:   db,
		Path: "",
	}
}

func NewWithPath(db *sql.DB, path string) *Repo {
	return &Repo{
		DB:   db,
		Path: path,
	}
}

func (r *Repo) Raw() *sql.DB {
	return r.DB
}

func (r *Repo) Close() error {
	if r.DB != nil {
		return r.DB.Close()
	}
	return nil
}
