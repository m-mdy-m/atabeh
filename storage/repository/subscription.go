package repository

import (
	"database/sql"
	"fmt"
)

func (r *Repo) AddSubscription(url string) error {
	_, err := r.core.DB.Exec("INSERT OR IGNORE INTO subscriptions (url) VALUES (?)", url)
	return err
}

func (r *Repo) ListSubscriptions() ([]string, error) {
	rows, err := r.core.DB.Query("SELECT url FROM subscriptions ORDER BY url")
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

func (r *Repo) SubscriptionExists(url string) (bool, error) {
	var exists int
	err := r.core.DB.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE url = ?", url).Scan(&exists)
	return exists > 0, err
}

func (r *Repo) GetLatestSubscription() (string, error) {
	var url string
	err := r.core.DB.QueryRow("SELECT url FROM subscriptions ORDER BY rowid DESC LIMIT 1").Scan(&url)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("no subscriptions found")
	}
	return url, err
}

func (r *Repo) RemoveSubscription(url string) error {
	res, err := r.core.DB.Exec("DELETE FROM subscriptions WHERE url = ?", url)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("subscription not found: %s", url)
	}
	return nil
}

func (r *Repo) ClearSubscriptions() error {
	_, err := r.core.DB.Exec("DELETE FROM subscriptions")
	return err
}
