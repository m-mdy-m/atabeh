package repository

import "github.com/m-mdy-m/atabeh/storage/core"

func (r *Repo) AddSubscription(url string) error {
	q := core.InsertInto(core.TableSubscriptions).
		Columns("url").
		Values(url).
		OrIgnore()

	_, err := r.core.InsertQuery(q)
	return err
}
func (r *Repo) ListSubscriptions() ([]string, error) {
	q := core.Select("url").
		From(core.TableSubscriptions).
		OrderBy("url")

	sqlStr, args := q.Build()
	rows, err := r.core.DB.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *Repo) RemoveSubscription(url string) error {
	q := core.DeleteFrom(core.TableSubscriptions).
		Where("url = ?", url)

	_, err := r.core.ExecQuery(q)
	return err
}
func (r *Repo) ClearSubscriptions() error {
	q := core.DeleteFrom(core.TableSubscriptions)
	_, err := r.core.ExecQuery(q)
	return err
}
