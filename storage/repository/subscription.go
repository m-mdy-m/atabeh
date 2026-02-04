package repository

import (
	"github.com/m-mdy-m/atabeh/storage/core"
)

func (r *Repo) AddSubscription(url string) error {
	q := core.InsertInto(core.TableSubscriptions).
		Columns(core.SubscriptionColURL).
		Values(url).
		OrIgnore()

	_, err := r.core.InsertQuery(q)
	return err
}

func (r *Repo) ListSubscriptions() ([]string, error) {
	q := core.Select(core.SubscriptionColURL).
		From(core.TableSubscriptions).
		OrderBy(core.SubscriptionColURL)

	sqlStr, args := q.Build()
	rows, err := r.core.DB.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, rows.Err()
}

func (r *Repo) RemoveSubscription(url string) error {
	q := core.DeleteFrom(core.TableSubscriptions).
		Where(core.SubscriptionColURL+" = ?", url)

	_, err := r.core.ExecQuery(q)
	return err
}

func (r *Repo) ClearSubscriptions() error {
	q := core.DeleteFrom(core.TableSubscriptions)
	_, err := r.core.ExecQuery(q)
	return err
}
