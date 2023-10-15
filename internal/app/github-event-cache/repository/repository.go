package repository

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

type Repository struct {
	db *sql.DB
}

type Event struct {
	Id        int64
	Type      string
	Repo      string
	Owner     string
	Payload   string
	User      string
	Timestamp int64
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS main.github_events
	(
	    type      TEXT    NOT NULL,
	    repo      TEXT    NOT NULL,
	    id        INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	    owner     TEXT    NOT NULL,
	    payload   TEXT    NOT NULL,
	    timestamp INTEGER NOT NULL,
	    user      TEXT    NOT NULL
	);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *Repository) SaveEvent(eventID int64, eventType, repo, owner, payload, user string, timestamp int64) error {
	// Check if event with the specific ID already exists
	var existingID int64
	err := r.db.QueryRow("SELECT id FROM main.github_events WHERE id = ?", eventID).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// If event with the ID already exists, we don't insert it again
	if existingID == eventID {
		return nil
	}

	query := `INSERT INTO main.github_events(id, type, repo, owner, payload, user, timestamp) VALUES(?, ?, ?, ?, ?, ?, ?)`
	_, err = r.db.Exec(query, eventID, eventType, repo, owner, payload, user, timestamp)
	return err
}

func (r *Repository) GetEvents(owner, repo string, startTimestamp, endTimestamp int64) ([]Event, error) {
	query := `
        SELECT id, type, repo, owner, payload, timestamp, user 
        FROM main.github_events 
        WHERE owner = ? AND repo = ? AND timestamp BETWEEN ? AND ?
    `
	rows, err := r.db.Query(query, owner, repo, startTimestamp, endTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		err = rows.Scan(&event.Id, &event.Type, &event.Repo, &event.Owner, &event.Payload, &event.Timestamp, &event.User)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
