-- name: GetSubsOfChannel :many
SELECT chat FROM subs
WHERE channel = $1;