-- name: Subscribe :one
INSERT INTO subs(chat, channel)
VALUES ($1, $2)
RETURNING *;

-- name: UnSubscribe :exec
DELETE FROM subs
WHERE chat = $1 AND channel = $2;

-- name: GetSubsOfChannel :many
SELECT chat FROM subs
WHERE channel = $1;

-- name: GroupIDChanged :exec
UPDATE subs SET chat = $2
WHERE chat = $1;