package migrations

import (
	"context"
	"database/sql"
	_ "modernc.org/sqlite"
	"testing"
)

func TestOverrideMigrationPreservesRowsAndIsIdempotent(t *testing.T) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	_, err = database.Exec("CREATE TABLE overrides (class_hash TEXT, is_hidden BOOLEAN); INSERT INTO overrides VALUES ('original', 1)")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := EnsureOverrideSource(context.Background(), database); err != nil {
			t.Fatal(err)
		}
	}
	var hash string
	var hidden bool
	var source int64
	if err := database.QueryRow("SELECT class_hash, is_hidden, source_column FROM overrides").Scan(&hash, &hidden, &source); err != nil {
		t.Fatal(err)
	}
	if hash != "original" || !hidden || source != 0 {
		t.Fatalf("lost override: %s %v %d", hash, hidden, source)
	}
}
