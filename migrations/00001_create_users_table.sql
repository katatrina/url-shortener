-- +goose Up
CREATE TABLE users
(
    "id"            uuid PRIMARY KEY,
    "email"         text UNIQUE NOT NULL,
    "full_name"     text,
    "password_hash" text        NOT NULL,
    "created_at"    timestamptz NOT NULL DEFAULT (now()),
    "updated_at"    timestamptz NOT NULL DEFAULT (now())
);

-- +goose Down
DROP TABLE users;