-- +goose Up

CREATE TABLE fetchers(
    id BIGINT PRIMARY KEY,
    api_id BIGINT,
    api_hash TEXT
);

-- +goose Down

DROP TABLE fetchers;