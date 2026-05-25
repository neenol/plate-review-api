-- name: CreateVote :one
INSERT INTO votes (user_id, target_type, target_id, value)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, target_type, target_id, value, created_at;

-- name: GetVote :one
SELECT id, user_id, target_type, target_id, value, created_at
FROM votes
WHERE user_id = $1 AND target_type = $2 AND target_id = $3;

-- name: DeleteVote :exec
DELETE FROM votes
WHERE user_id = $1 AND target_type = $2 AND target_id = $3;
