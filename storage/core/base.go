package core

import (
	"database/sql"
)

type Repo struct {
	DB  *sql.DB
	Path string
}

func New(db *sql.DB, path string) Repo {
	return Repo{DB: db, Path: path}
}
