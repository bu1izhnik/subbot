-- name: Subscribe :one
INSERT INTO subs(chat, channel, thread)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UnSubscribe :exec
DELETE FROM subs
WHERE chat = $1 AND channel = $2;

-- name: GetSubsOfChannel :many
SELECT chat, thread FROM subs
WHERE channel = $1;

-- name: GroupIDChanged :exec
UPDATE subs SET chat = $2
WHERE chat = $1;

-- name: CountSubsOfChannel :one
SELECT COUNT(*) FROM subs
WHERE channel = $1;

-- name: CheckSub :one
SELECT COUNT(1) FROM subs
WHERE chat = $1 AND channel = $2;