BEGIN;

CREATE TABLE links
(
    id              UUID PRIMARY KEY,
    user_id         UUID REFERENCES users (id),
    destination_url TEXT        NOT NULL,
    short_code      TEXT        NOT NULL,

    click_count     BIGINT      NOT NULL DEFAULT 0,
    expires_at      TIMESTAMPTZ,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP
);

CREATE UNIQUE INDEX links_short_code_unique
    ON links (short_code)
    WHERE deleted_at IS NULL;

COMMIT;