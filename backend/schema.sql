CREATE TABLE IF NOT EXISTS users (
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

CREATE TABLE IF NOT EXISTS overrides (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id        INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    class_hash     TEXT    NOT NULL,
    start_hour     TEXT    NOT NULL DEFAULT '',
    is_hidden      BOOLEAN NOT NULL DEFAULT 0,
    is_transparent BOOLEAN NOT NULL DEFAULT 0,
    UNIQUE(user_id, class_hash)
);

CREATE TABLE IF NOT EXISTS buildings (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    key             TEXT UNIQUE,
    google_place_id TEXT UNIQUE,
    name            TEXT,
    address         TEXT
);

INSERT INTO buildings (key, google_place_id) VALUES
    ('ARCH', 'ChIJVanSewNbFkcRis321uI90wQ'), -- Architektury
    ('CHEM', 'ChIJD6D2gARbFkcROkHrxYrK1Xc'), -- Chemiczny
    ('DZAL', 'ChIJ26n3dQRbFkcRcSq8i7pOAxY'), -- Działownia
    ('HOUS', 'ChIJF7aaXgBbFkcRlAzBvmSHoRc'), -- Houston
    ('GGIL', 'ChIJoYVxbQNbFkcRLVeU6GMG0ZM'), -- Galeria GIL
    ('SJOP', 'ChIJ5bx6WwNbFkcREw-iUfxzr_8')  -- Studium języków obcych
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS classrooms (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE
);

INSERT INTO classrooms (name) VALUES
    ('114 GIL'),
    ('131'),
    ('135'),
    ('136'),
    ('138'),
    ('149b'),
    ('150'),
    ('151'),
    ('152'),
    ('N'),
    ('Działownia 1/15'),
    ('wykładowa KI'),
    ('S1'),
    ('S2'),
    ('S3')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS building_classrooms (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    building_id     INTEGER NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    classroom_id INTEGER NOT NULL REFERENCES classrooms(id) ON DELETE CASCADE
);

INSERT INTO building_classrooms (building_id, classroom_id) VALUES
    ((SELECT id FROM buildings WHERE key = 'GGIL'), (SELECT id FROM classrooms WHERE name = '114 GIL')),
    ((SELECT id FROM buildings WHERE key = 'HOUS'), (SELECT id FROM classrooms WHERE name = '131')),
    ((SELECT id FROM buildings WHERE key = 'CHEM'), (SELECT id FROM classrooms WHERE name = '135')),
    ((SELECT id FROM buildings WHERE key = 'CHEM'), (SELECT id FROM classrooms WHERE name = '136')),
    ((SELECT id FROM buildings WHERE key = 'SJOP'), (SELECT id FROM classrooms WHERE name = '138')),
    ((SELECT id FROM buildings WHERE key = 'SJOP'), (SELECT id FROM classrooms WHERE name = '149b')),
    ((SELECT id FROM buildings WHERE key = 'SJOP'), (SELECT id FROM classrooms WHERE name = '150')),
    ((SELECT id FROM buildings WHERE key = 'CHEM'), (SELECT id FROM classrooms WHERE name = '151')),
    ((SELECT id FROM buildings WHERE key = 'SJOP'), (SELECT id FROM classrooms WHERE name = '152')),
    ((SELECT id FROM buildings WHERE key = 'ARCH'), (SELECT id FROM classrooms WHERE name = 'N')),
    ((SELECT id FROM buildings WHERE key = 'DZAL'), (SELECT id FROM classrooms WHERE name = 'Działownia 1/15')),
    ((SELECT id FROM buildings WHERE key = 'CHEM'), (SELECT id FROM classrooms WHERE name = 'wykładowa KI')),
    ((SELECT id FROM buildings WHERE key = 'GGIL'), (SELECT id FROM classrooms WHERE name = 'S1')),
    ((SELECT id FROM buildings WHERE key = 'GGIL'), (SELECT id FROM classrooms WHERE name = 'S2')),
    ((SELECT id FROM buildings WHERE key = 'GGIL'), (SELECT id FROM classrooms WHERE name = 'S3'))
ON CONFLICT DO NOTHING;
