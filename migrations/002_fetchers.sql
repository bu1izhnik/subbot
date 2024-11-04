-- +goose Up

CREATE TABLE fetchers(
    id BIGINT PRIMARY KEY,
    phone TEXT,
    ip TEXT,
    port TEXT
);

-- +goose Down

DROP TABLE fetchers;