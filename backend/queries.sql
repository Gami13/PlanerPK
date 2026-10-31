-- name: UpsertUser :one
INSERT INTO users (email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in)
VALUES (sqlc.arg(email), sqlc.arg(name), sqlc.arg(given_name), sqlc.arg(picture), sqlc.arg(access_token), sqlc.arg(token_type), sqlc.arg(refresh_token), sqlc.arg(expiry), sqlc.arg(expires_in))
ON CONFLICT(email) DO UPDATE SET
    name          = excluded.name,
    given_name    = excluded.given_name,
    picture       = excluded.picture,
    access_token  = excluded.access_token,
    token_type    = excluded.token_type,
    refresh_token = excluded.refresh_token,
    expiry        = excluded.expiry,
    expires_in    = excluded.expires_in
RETURNING id, email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in, chosen_column, calendar_id, is_active, last_synced;

-- name: GetUserByEmail :one
SELECT id, email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in, chosen_column, calendar_id, is_active, last_synced
FROM users
WHERE email = sqlc.arg(email);

-- name: GetAllUsers :many
SELECT id, email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in, chosen_column, calendar_id, is_active, last_synced
FROM users;

-- name: UpdateChosenColumn :exec
UPDATE users SET chosen_column = sqlc.arg(chosen_column) WHERE access_token = sqlc.arg(access_token);

-- name: UpdateUserActive :exec
UPDATE users SET is_active = sqlc.arg(is_active) WHERE access_token = sqlc.arg(access_token);

-- name: GetTokenByAccessToken :one
SELECT access_token, token_type, refresh_token, expiry, expires_in FROM users WHERE access_token = sqlc.arg(access_token);

-- name: GetUserByAccessToken :one
SELECT id, email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in, chosen_column, calendar_id, is_active, last_synced
FROM users WHERE access_token = sqlc.arg(access_token);

-- name: UpdateUserSyncTime :exec
UPDATE users SET last_synced = (strftime('%Y-%m-%d %H:%M:%S+00:00', 'now')) WHERE access_token = sqlc.arg(access_token);

-- name: UpdateUserCalendarID :exec
UPDATE users SET calendar_id = sqlc.arg(calendar_id) WHERE id = sqlc.arg(id);

-- name: GetOverridesForUser :many
SELECT id, user_id, class_hash, start_hour, is_hidden, is_transparent, source_column FROM overrides WHERE user_id = sqlc.arg(user_id);

-- name: UpsertOverride :exec
INSERT INTO overrides (user_id, class_hash, start_hour, is_hidden, is_transparent, source_column)
VALUES (sqlc.arg(user_id), sqlc.arg(class_hash), sqlc.arg(start_hour), sqlc.arg(is_hidden), sqlc.arg(is_transparent), sqlc.arg(source_column))
ON CONFLICT(user_id, class_hash) DO UPDATE SET start_hour = excluded.start_hour, is_hidden = excluded.is_hidden, is_transparent = excluded.is_transparent, source_column = excluded.source_column;

-- name: DeleteOverride :exec
DELETE FROM overrides WHERE user_id = sqlc.arg(user_id) AND class_hash = sqlc.arg(class_hash);

-- name: UpsertBuildingPlace :exec
INSERT INTO building_places (building_key, canonical_name)
VALUES (sqlc.arg(building_key), sqlc.arg(canonical_name))
ON CONFLICT(building_key) DO UPDATE SET canonical_name = excluded.canonical_name;

-- name: GetBuildingPlaces :many
SELECT building_key, canonical_name, place_id, display_name, address, confidence, verified_at, expires_at
FROM building_places ORDER BY building_key;

-- name: UpdateBuildingPlace :exec
UPDATE building_places
SET place_id = sqlc.arg(place_id), display_name = sqlc.arg(display_name), address = sqlc.arg(address),
    confidence = sqlc.arg(confidence), verified_at = sqlc.arg(verified_at), expires_at = sqlc.arg(expires_at)
WHERE building_key = sqlc.arg(building_key);
