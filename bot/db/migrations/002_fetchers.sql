-- +goose Up

CREATE TABLE fetchers(
    id BIGINT PRIMARY KEY,
    phone TEXT NOT NULL,
    ip TEXT NOT NULL,
    port TEXT NOT NULL
);

-- +goose Down

DROP TABLE fetchers;