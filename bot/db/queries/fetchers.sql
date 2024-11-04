-- name: CheckFetcher :one
SELECT COUNT(1)
FROM fetchers
WHERE id = $1;