-- name: GetPlateByHash :one
SELECT id, plate_hash, state, created_at
FROM plates
WHERE plate_hash = $1;

-- name: CreatePlate :one
INSERT INTO plates (plate_hash, state)
VALUES ($1, $2)
RETURNING id, plate_hash, state, created_at;
