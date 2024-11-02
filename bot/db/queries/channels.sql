-- name: ChangeChannelUsername :exec
UPDATE channels
SET username = $2
WHERE id = $1;

-- name: GetUsernamesOfGroupSubs :many
SELECT channels.username FROM channels
JOIN subs ON channels.id = subs.channel
WHERE subs.chat = $1;