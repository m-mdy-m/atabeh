package repository

import (
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/core"
)

type Repo struct {
	core *core.Repo
	db   *storage.RepoDatabase
}

func New(coreRepo *core.Repo) *Repo {
	return &Repo{
		core: coreRepo,
		db:   nil,
	}
}
func NewFromDB(db *storage.RepoDatabase) *Repo {
	return &Repo{
		core: db.Repo,
		db:   db,
	}
}
