-- +goose Up
CREATE TABLE "links"
(
    "id"              uuid PRIMARY KEY,
    "user_id"         uuid        NOT NULL,
    "slug"            text UNIQUE NOT NULL,
    "destination_url" text        NOT NULL,
    "title"           text        NOT NULL DEFAULT '',
    "is_custom_slug"  boolean     NOT NULL DEFAULT false,
    "created_at"      timestamptz NOT NULL DEFAULT (now()),
    "updated_at"      timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "links"
    ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;

-- +goose Down
DROP TABLE "links";