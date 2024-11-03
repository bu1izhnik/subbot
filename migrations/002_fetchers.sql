-- +goose Up

CREATE TABLE fetchers(
    id BIGINT PRIMARY KEY,
    phone TEXT
);

-- +goose Down

DROP TABLE fetchers;