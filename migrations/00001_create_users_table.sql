-- +goose Up
CREATE TABLE users
(
    "id"            uuid PRIMARY KEY,
    "email"         text UNIQUE NOT NULL,
    "full_name"     text        NOT NULL DEFAULT '',
    "password_hash" text,
    "created_at"    timestamptz NOT NULL DEFAULT (now()),
    "updated_at"    timestamptz NOT NULL DEFAULT (now())
);

-- +goose Down
DROP TABLE users;