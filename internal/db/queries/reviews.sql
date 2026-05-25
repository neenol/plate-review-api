-- name: CreateReview :one
INSERT INTO reviews (plate_id, user_id, body)
VALUES ($1, $2, $3)
RETURNING id, plate_id, user_id, body, created_at;

-- name: GetReviewByID :one
SELECT r.id, r.plate_id, r.user_id, r.body, r.created_at,
       u.username AS author_username,
       COALESCE(SUM(v.value), 0)::bigint AS score
FROM reviews r
JOIN users u ON u.id = r.user_id
LEFT JOIN votes v ON v.target_id = r.id AND v.target_type = 'review'
WHERE r.id = $1
GROUP BY r.id, r.plate_id, r.user_id, r.body, r.created_at, u.username;

-- name: ListReviewsByPlate :many
SELECT r.id, r.plate_id, r.user_id, r.body, r.created_at,
       u.username AS author_username,
       COALESCE(SUM(v.value), 0)::bigint AS score
FROM reviews r
JOIN users u ON u.id = r.user_id
LEFT JOIN votes v ON v.target_id = r.id AND v.target_type = 'review'
WHERE r.plate_id = $1
GROUP BY r.id, r.plate_id, r.user_id, r.body, r.created_at, u.username
ORDER BY r.created_at DESC;

-- name: DeleteReview :exec
DELETE FROM reviews WHERE id = $1;

-- name: GetReviewOwnerID :one
SELECT user_id FROM reviews WHERE id = $1;
