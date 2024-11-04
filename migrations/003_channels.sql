-- +goose Up

CREATE TABLE channels(
    id BIGINT PRIMARY KEY,
    hash BIGINT NOT NULL,
    username TEXT NOT NULL,
    stored_at BIGINT NOT NULL
);

-- +goose Down

DROP TABLE channels;