package repository

import (
	"github.com/m-mdy-m/atabeh/storage/core"
)

/* GLobal */
type Repo struct {
	core *core.Repo
}

func New(coreRepo *core.Repo) *Repo {
	return &Repo{core: coreRepo}
}
