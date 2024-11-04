-- name: AddFetcher :one
INSERT INTO fetchers(id, phone)
VALUES ($1, $2)
RETURNING *;

-- name: DeleteFetcher :exec
DELETE FROM fetchers
WHERE id = $1;

-- name: GetLeastFullFetcher :one
SELECT fetchers.id, fetchers.ip, fetchers.port
FROM fetchers JOIN channels
ON fetchers.id = channels.stored_at
GROUP BY fetchers.id
ORDER BY COUNT(*) ASC
LIMIT 1;

-- name: CheckFetcher :one
SELECT COUNT(1)
FROM fetchers
WHERE id = $1;