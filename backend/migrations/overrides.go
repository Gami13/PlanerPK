package migrations

import (
	"context"
	"database/sql"
)

// EnsureOverrideSource upgrades existing databases without deleting saved overrides.
func EnsureOverrideSource(ctx context.Context, database *sql.DB) error {
	rows, err := database.QueryContext(ctx, "PRAGMA table_info(overrides)")
	if err != nil {
		return err
	}
	found := false
	for rows.Next() {
		var cid, notNull, pk int
		var name, kind string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &kind, &notNull, &defaultValue, &pk); err != nil {
			rows.Close()
			return err
		}
		if name == "source_column" {
			found = true
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil || found {
		return err
	}
	_, err = database.ExecContext(ctx, "ALTER TABLE overrides ADD COLUMN source_column INTEGER NOT NULL DEFAULT 0")
	return err
}
