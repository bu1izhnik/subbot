-- name: AddFetcher :one
INSERT INTO fetchers(id, phone)
VALUES ($1, $2)
RETURNING *;

-- name: DeleteFetcher :exec
DELETE FROM fetchers
WHERE id = $1;

-- name: GetEmptyFetcher :one
SELECT id FROM fetchers
ORDER BY RANDOM()
LIMIT 1;