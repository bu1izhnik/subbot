-- name: AddChannel :one
INSERT INTO channels(id, hash, username, stored_at)
VALUES($1, $2, $3, $4)
RETURNING *;

-- name: DeleteChannel :exec
DELETE FROM channels
WHERE id = $1;

-- name: ChangeChannelUsername :exec
UPDATE channels
SET username = $2
WHERE id = $1;

-- name: ChangeChannelUsernameAndHash :exec
UPDATE channels
SET username = $2, hash = $3
WHERE id = $1;

-- name: GetUsernamesOfGroupSubs :many
SELECT channels.username FROM channels
JOIN subs ON channels.id = subs.channel
WHERE subs.chat = $1;