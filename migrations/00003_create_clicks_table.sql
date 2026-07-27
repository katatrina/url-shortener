-- +goose Up
-- Append-only click event log
CREATE TABLE "clicks"
(
    "id"           uuid PRIMARY KEY,
    "link_id"      uuid        NOT NULL,
    "clicked_at"   timestamptz NOT NULL,
    "ip_address"   inet,
    "referrer"     text,
    "country_code" text
);

ALTER TABLE "clicks"
    ADD FOREIGN KEY ("link_id") REFERENCES "links" ("id") ON DELETE CASCADE;

-- +goose Down
DROP TABLE "clicks";
