BEGIN TRANSACTION;

CREATE TABLE users_new (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    email         TEXT UNIQUE,
    name          TEXT,
    given_name    TEXT,
    picture       TEXT,
    access_token  TEXT    NOT NULL,
    token_type    TEXT    NOT NULL,
    refresh_token TEXT    NOT NULL,
    expiry        TIMESTAMP NOT NULL,
    expires_in    INTEGER NOT NULL,
    chosen_column INTEGER DEFAULT -1,
    is_active     BOOLEAN DEFAULT 1,
    last_synced   TIMESTAMP NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S+00:00', 'now'))
);

INSERT INTO users_new SELECT * FROM users;

DROP TABLE users;

ALTER TABLE users_new RENAME TO users;

COMMIT;
