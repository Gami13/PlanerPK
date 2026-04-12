-- migration2.sql: adds the overrides table for existing databases
-- Safe to run multiple times due to IF NOT EXISTS.

CREATE TABLE IF NOT EXISTS overrides (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id        INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    class_hash     TEXT    NOT NULL,
    start_hour     TEXT    NOT NULL DEFAULT '',
    is_hidden      BOOLEAN NOT NULL DEFAULT 0,
    is_transparent BOOLEAN NOT NULL DEFAULT 0,
    UNIQUE(user_id, class_hash)
);
