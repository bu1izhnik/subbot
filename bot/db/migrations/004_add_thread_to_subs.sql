-- +goose Up

ALTER TABLE subs
ADD COLUMN thread INTEGER NOT NULL DEFAULT 0;

-- +goose Down

ALTER TABLE subs
DROP COLUMN thread;