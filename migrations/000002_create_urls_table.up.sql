BEGIN;

CREATE TABLE urls
(
    id           UUID PRIMARY KEY,
    short_code   TEXT        NOT NULL,
    original_url TEXT        NOT NULL,
    user_id      UUID REFERENCES users (id),

    click_count  BIGINT      NOT NULL DEFAULT 0,

    expires_at   TIMESTAMPTZ,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX urls_short_code_unique
    ON urls (short_code)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_urls_user_id
    ON urls (user_id, created_at DESC)
    WHERE deleted_at IS NULL;

COMMIT;