-- name: GetUserByID :one
SELECT id, clerk_user_id, username, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByClerkID :one
SELECT id, clerk_user_id, username, created_at, updated_at
FROM users
WHERE clerk_user_id = $1;

-- name: GetUserByUsername :one
SELECT id, clerk_user_id, username, created_at, updated_at
FROM users
WHERE username = $1;

-- name: CreateUser :one
INSERT INTO users (clerk_user_id, username)
VALUES ($1, $2)
RETURNING id, clerk_user_id, username, created_at, updated_at;

-- name: UpdateUserUsername :one
UPDATE users
SET username = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, clerk_user_id, username, created_at, updated_at;
