-- +goose Up

CREATE TABLE subs(
    chat BIGINT NOT NULL,
    channel BIGINT NOT NULL,
    UNIQUE(chat, channel)
);

-- +goose Down

DROP TABLE subs;
