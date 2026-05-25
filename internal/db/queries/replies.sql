-- name: CreateReply :one
INSERT INTO replies (review_id, parent_reply_id, user_id, body)
VALUES ($1, $2, $3, $4)
RETURNING id, review_id, parent_reply_id, user_id, body, created_at;

-- name: GetReplyByID :one
SELECT r.id, r.review_id, r.parent_reply_id, r.user_id, r.body, r.created_at,
       u.username AS author_username,
       COALESCE(SUM(v.value), 0)::bigint AS score
FROM replies r
JOIN users u ON u.id = r.user_id
LEFT JOIN votes v ON v.target_id = r.id AND v.target_type = 'reply'
WHERE r.id = $1
GROUP BY r.id, r.review_id, r.parent_reply_id, r.user_id, r.body, r.created_at, u.username;

-- name: ListRepliesByReview :many
SELECT r.id, r.review_id, r.parent_reply_id, r.user_id, r.body, r.created_at,
       u.username AS author_username,
       COALESCE(SUM(v.value), 0)::bigint AS score
FROM replies r
JOIN users u ON u.id = r.user_id
LEFT JOIN votes v ON v.target_id = r.id AND v.target_type = 'reply'
WHERE r.review_id = $1 AND r.parent_reply_id IS NULL
GROUP BY r.id, r.review_id, r.parent_reply_id, r.user_id, r.body, r.created_at, u.username
ORDER BY r.created_at ASC;

-- name: ListRepliesByParent :many
SELECT r.id, r.review_id, r.parent_reply_id, r.user_id, r.body, r.created_at,
       u.username AS author_username,
       COALESCE(SUM(v.value), 0)::bigint AS score
FROM replies r
JOIN users u ON u.id = r.user_id
LEFT JOIN votes v ON v.target_id = r.id AND v.target_type = 'reply'
WHERE r.parent_reply_id = $1
GROUP BY r.id, r.review_id, r.parent_reply_id, r.user_id, r.body, r.created_at, u.username
ORDER BY r.created_at ASC;

-- name: DeleteReply :exec
DELETE FROM replies WHERE id = $1;

-- name: GetReplyOwnerID :one
SELECT user_id FROM replies WHERE id = $1;
