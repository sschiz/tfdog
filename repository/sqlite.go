package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// sqliteRepo is sqlite implementation of Repository
type sqliteRepo struct {
	db *sql.DB
}

// NewSqliteRepository returns new Repository instance.
func NewSqliteRepository(dsn string) (Repository, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return &sqliteRepo{db: db}, nil
}

func (s *sqliteRepo) SaveSubscription(sub Subscription) error {
	const query = `
INSERT INTO subscriptions (user_id, app_name, link)
SELECT :user_id, :app_name, :link
WHERE NOT EXISTS(SELECT 1 FROM subscriptions WHERE user_id = :user_id AND link = :link);
`
	_, err := s.db.Exec(
		query,
		sql.Named("user_id", sub.UserID),
		sql.Named("app_name", sub.AppName),
		sql.Named("link", sub.Link),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *sqliteRepo) RemoveSubscription(sub Subscription) error {
	const query = `DELETE FROM subscriptions WHERE user_id = ? AND link = ?`
	_, err := s.db.Exec(query, sub.UserID, sub.Link)
	if err != nil {
		return err
	}

	return nil
}

func (s *sqliteRepo) GetUserSubscriptions(userID int) ([]Subscription, error) {
	const query = `SELECT user_id, app_name, link FROM subscriptions WHERE user_id = ?`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var res []Subscription
	for rows.Next() {
		var sub Subscription
		err = rows.Scan(&sub.UserID, &sub.AppName, &sub.Link)
		if err != nil {
			return nil, err
		}
		res = append(res, sub)
	}

	return res, nil
}

func (s *sqliteRepo) GetAllSubscriptions() ([]Subscription, error) {
	const query = `SELECT user_id, app_name, link FROM subscriptions`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var res []Subscription
	for rows.Next() {
		var sub Subscription
		err = rows.Scan(&sub.UserID, &sub.AppName, &sub.Link)
		if err != nil {
			return nil, err
		}
		res = append(res, sub)
	}

	return res, nil
}

func (s *sqliteRepo) Close() error {
	return s.db.Close()
}
