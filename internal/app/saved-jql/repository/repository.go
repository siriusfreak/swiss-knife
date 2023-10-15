package repository

import (
	"context"
	"database/sql"
)

type SavedJQL struct {
	ID   string
	Name string
	JQL  string
}

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTables(ctx context.Context) error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS saved_jql (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		jql TEXT NOT NULL
	);`
	_, err := r.db.ExecContext(ctx, createTableSQL)
	return err
}

func (r *Repository) SaveJQL(ctx context.Context, jql *SavedJQL) error {
	statement, err := r.db.PrepareContext(ctx, "INSERT INTO saved_jql (name, jql) VALUES (?, ?)")
	if err != nil {
		return err
	}
	_, err = statement.ExecContext(ctx, jql.Name, jql.JQL)
	return err
}

func (r *Repository) GetAllJQL(ctx context.Context) ([]*SavedJQL, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, jql FROM saved_jql")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jqls []*SavedJQL
	for rows.Next() {
		var jql SavedJQL
		if err := rows.Scan(&jql.ID, &jql.Name, &jql.JQL); err != nil {
			return nil, err
		}
		jqls = append(jqls, &jql)
	}
	return jqls, nil
}

func (r *Repository) DeleteJQL(ctx context.Context, id string) error {
	statement, err := r.db.PrepareContext(ctx, "DELETE FROM saved_jql WHERE id = ?")
	if err != nil {
		return err
	}
	_, err = statement.ExecContext(ctx, id)
	return err
}
