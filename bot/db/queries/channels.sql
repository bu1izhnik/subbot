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

-- name: GetGroupSubs :many
SELECT channels.id, channels.username FROM channels
JOIN subs ON channels.id = subs.channel
WHERE subs.chat = $1;

-- name: ChannelAlreadyStored :one
SELECT COUNT(1)
FROM channels
WHERE id = $1;

-- name: GetChannelsFetcher :one
SELECT fetchers.ip, fetchers.port FROM fetchers
JOIN channels ON channels.stored_at = fetchers.id
WHERE channels.id = $1;