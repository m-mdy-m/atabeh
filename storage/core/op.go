package core

import "database/sql"

func (r Repo) InsertQuery(q *Query) (int64, error) {
	sqlStr, args := q.Build()
	res, err := r.DB.Exec(sqlStr, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r Repo) ExecQuery(q *Query) (int64, error) {
	sqlStr, args := q.Build()
	res, err := r.DB.Exec(sqlStr, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func GetOne[T any](row *sql.Row, scan func(Scanner) (*T, error)) (*T, error) {
	obj, err := scan(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return obj, err
}

func List[T any](rows *sql.Rows, scan func(Scanner) (*T, error)) ([]*T, error) {
	var out []*T
	for rows.Next() {
		obj, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, obj)
	}
	return out, rows.Err()
}

func WithTx(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
