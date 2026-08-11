CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE,
    name TEXT,
    given_name TEXT,
    picture TEXT,
    access_token TEXT NOT NULL,
    token_type TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expiry TIMESTAMP NOT NULL,
    expires_in INTEGER NOT NULL,
    chosen_column INTEGER DEFAULT -1,
    is_active BOOLEAN DEFAULT 1,
    last_synced TIMESTAMP NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S+00:00', 'now'))
);

CREATE TABLE IF NOT EXISTS overrides (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    class_hash TEXT NOT NULL,
    start_hour TEXT NOT NULL DEFAULT '',
    is_hidden BOOLEAN NOT NULL DEFAULT 0,
    is_transparent BOOLEAN NOT NULL DEFAULT 0,
    UNIQUE(user_id, class_hash)
);

-- This is intentionally a fresh schema. Remove the old state/database.db before
-- deploying it; the old seeded building tables are not migrated.
CREATE TABLE IF NOT EXISTS building_places (
    building_key TEXT PRIMARY KEY,
    canonical_name TEXT NOT NULL,
    place_id TEXT UNIQUE,
    display_name TEXT,
    address TEXT,
    confidence REAL NOT NULL DEFAULT 0,
    verified_at TIMESTAMP,
    expires_at TIMESTAMP
);
