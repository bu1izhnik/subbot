-- name: AddChannel :one
INSERT INTO channels(id, hash, username, stored_at)
VALUES($1, $2, $3, $4)
RETURNING *;

-- name: DeleteChannel :exec
DELETE FROM channels
WHERE id = $1;