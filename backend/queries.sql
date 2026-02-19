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
UPDATE users
SET chosen_column = sqlc.arg(chosen_column)
WHERE access_token = sqlc.arg(access_token);

-- name: GetTokenByAccessToken :one
SELECT access_token, token_type, refresh_token, expiry, expires_in
FROM users
WHERE access_token = sqlc.arg(access_token);

-- name: GetUserByAccessToken :one
SELECT id, email, name, given_name, picture, access_token, token_type, refresh_token, expiry, expires_in, chosen_column, is_active, last_synced
FROM users
WHERE access_token = sqlc.arg(access_token);

-- name: UpdateUserSyncTime :exec
UPDATE users
SET last_synced = (strftime('%Y-%m-%d %H:%M:%S+00:00', 'now'))
WHERE access_token = sqlc.arg(access_token);
