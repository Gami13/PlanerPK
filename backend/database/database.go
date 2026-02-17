package database

import (
	"database/sql"
	_ "embed"

	"github.com/WoodPKrs/GoogleCalendarTimetable/database/gen"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

// Queries wraps the sqlc-generated gen.Queries and maps results to domain types.
type Queries struct {
	*gen.Queries
}

func New(db *sql.DB) *Queries {
	return &Queries{gen.New(db)}
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{q.Queries.WithTx(tx)}
}

// InitializeDatabase creates tables if they don't exist.
func InitializeDatabase(db *sql.DB) error {
	_, err := db.Exec(ddl)
	return err
}
