package repository

import (
	"github.com/m-mdy-m/atabeh/storage"
	"github.com/m-mdy-m/atabeh/storage/core"
)

/* GLobal */
type Repo struct {
	core *core.Repo
}

/*
   =========
   PROFILE
   =========
*/

type ProfileRepo interface {
	CreateProfile(name, source, profileType string) (int64, error)
	GetOrCreateProfile(name, source, profileType string) (int64, error)
	ListProfiles() ([]*storage.ProfileRow, error)
	GetProfile(id int) (*storage.ProfileRow, error)
	UpdateProfileSyncTime(id int64) error
	DeleteProfile(id int) error
}

func NewProfileRepo(r *core.Repo) ProfileRepo {
	return &Repo{core: r}
}

/*
   ===============
   SUBSCRIPTION
   ===============
*/

type SubscriptionRepo interface {
	AddSubscription(url string) error
	ListSubscriptions() ([]string, error)
	RemoveSubscription(url string) error
	ClearSubscriptions() error
}

func NewSubscriptionRepo(r *core.Repo) SubscriptionRepo {
	return &Repo{core: r}
}
