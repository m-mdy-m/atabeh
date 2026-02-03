package storage

import "database/sql"

type DB struct {
	inner *sql.DB
	Path  string
}
type ConfigRepo struct {
	db *DB
}
