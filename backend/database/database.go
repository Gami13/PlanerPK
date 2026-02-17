package database

import (
	"database/sql"
	_ "embed"

	"github.com/WoodPKrs/GoogleCalendarTimetable/gen/db"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

// Queries wraps the sqlc-generated db.Queries and maps results to domain types.
type Queries struct {
	*db.Queries
}

func New(d *sql.DB) *Queries {
	return &Queries{db.New(d)}
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{q.Queries.WithTx(tx)}
}

// InitializeDatabase creates tables if they don't exist.
func InitializeDatabase(db *sql.DB) error {
	_, err := db.Exec(ddl)
	return err
}
