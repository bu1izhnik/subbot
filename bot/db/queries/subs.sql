-- name: Subscribe :one
INSERT INTO subs(chat, channel)
VALUES ($1, $2)
RETURNING *;

-- name: UnSubscribe :exec
DELETE FROM subs
WHERE chat = $1 AND channel = $2;

-- name: ListGroupSubs :many
SELECT channel FROM subs
WHERE chat = $1;

-- name: GetSubsOfChannel :many
SELECT chat FROM subs
WHERE channel = $1;

-- name: CheckSubscription :one
SELECT COUNT(1) FROM subs
WHERE chat = $1 AND channel = $2;