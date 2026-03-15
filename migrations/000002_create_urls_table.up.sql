BEGIN;

CREATE TABLE urls
(
    -- Identity
    id           UUID PRIMARY KEY,
    short_code   TEXT        NOT NULL,
    original_url TEXT        NOT NULL,
    user_id      UUID REFERENCES users (id),

    -- Stats (denormalized, will be replaced by the analytics pipeline later)
    click_count  BIGINT      NOT NULL DEFAULT 0,

    -- Expiry
    expires_at   TIMESTAMPTZ,

    -- Timestamps
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,

    CONSTRAINT urls_short_code_unique UNIQUE (short_code)
);

-- Redirect lookup: the hottest endpoint, must be fast
CREATE INDEX idx_urls_short_code
    ON urls (short_code)
    WHERE deleted_at IS NULL;

-- User's URLs listing
CREATE INDEX idx_urls_user_id
    ON urls (user_id, created_at DESC)
    WHERE deleted_at IS NULL;

COMMIT;