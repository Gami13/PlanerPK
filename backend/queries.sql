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
RETURNING id, email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in, chosen_column, is_active, last_synced;

-- name: GetUserByEmail :one
SELECT id, email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in, chosen_column, is_active, last_synced
FROM users
WHERE email = sqlc.arg(email);

-- name: GetAllUsers :many
SELECT id, email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in, chosen_column, is_active, last_synced
FROM users;

-- name: UpdateChosenColumn :exec
UPDATE users SET chosen_column = sqlc.arg(chosen_column) WHERE access_token = sqlc.arg(access_token);

-- name: UpdateUserActive :exec
UPDATE users SET is_active = sqlc.arg(is_active) WHERE access_token = sqlc.arg(access_token);

-- name: GetTokenByAccessToken :one
SELECT access_token, token_type, refresh_token, expiry, expires_in FROM users WHERE access_token = sqlc.arg(access_token);

-- name: GetUserByAccessToken :one
SELECT id, email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in, chosen_column, is_active, last_synced
FROM users WHERE access_token = sqlc.arg(access_token);

-- name: UpdateUserSyncTime :exec
UPDATE users SET last_synced = (strftime('%Y-%m-%d %H:%M:%S+00:00', 'now')) WHERE access_token = sqlc.arg(access_token);

-- name: GetOverridesForUser :many
SELECT id, user_id, class_hash, start_hour, is_hidden, is_transparent FROM overrides WHERE user_id = sqlc.arg(user_id);

-- name: UpsertOverride :exec
INSERT INTO overrides (user_id, class_hash, start_hour, is_hidden, is_transparent)
VALUES (sqlc.arg(user_id), sqlc.arg(class_hash), sqlc.arg(start_hour), sqlc.arg(is_hidden), sqlc.arg(is_transparent))
ON CONFLICT(user_id, class_hash) DO UPDATE SET start_hour = excluded.start_hour, is_hidden = excluded.is_hidden, is_transparent = excluded.is_transparent;

-- name: DeleteOverride :exec
DELETE FROM overrides WHERE user_id = sqlc.arg(user_id) AND class_hash = sqlc.arg(class_hash);

-- name: UpsertBuildingPlace :exec
INSERT INTO building_places (building_key, canonical_name)
VALUES (sqlc.arg(building_key), sqlc.arg(canonical_name))
ON CONFLICT(building_key) DO UPDATE SET canonical_name = excluded.canonical_name;

-- name: GetBuildingPlaces :many
SELECT building_key, canonical_name, place_id, display_name, address, confidence, verified_at, expires_at
FROM building_places ORDER BY building_key;

-- name: UpsertClassroomBuilding :exec
INSERT INTO classroom_buildings (classroom_key, building_key)
VALUES (sqlc.arg(classroom_key), sqlc.arg(building_key))
ON CONFLICT(classroom_key) DO UPDATE SET building_key = excluded.building_key;

-- name: ClearClassroomBuildings :exec
DELETE FROM classroom_buildings;

-- name: UpdateBuildingPlace :exec
UPDATE building_places
SET place_id = sqlc.arg(place_id), display_name = sqlc.arg(display_name), address = sqlc.arg(address),
    confidence = sqlc.arg(confidence), verified_at = sqlc.arg(verified_at), expires_at = sqlc.arg(expires_at)
WHERE building_key = sqlc.arg(building_key);

-- name: ClearBuildingPlace :exec
UPDATE building_places
SET place_id = NULL, display_name = NULL, address = NULL, confidence = 0, verified_at = NULL, expires_at = NULL
WHERE building_key = sqlc.arg(building_key);
