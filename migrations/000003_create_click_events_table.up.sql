BEGIN;

CREATE TABLE click_events
(
    id         UUID PRIMARY KEY,
    url_id     UUID        NOT NULL REFERENCES urls (id),
    ip_address INET,
    user_agent TEXT,
    referer    TEXT,
    country    TEXT,
    clicked_at TIMESTAMPTZ NOT NULL
);

-- Analytics queries always filter by url_id first, then time range.
-- This index covers: "give me all clicks for URL X in the last 7 days"
CREATE INDEX idx_click_events_url_id_clicked_at
    ON click_events (url_id, clicked_at DESC);

CREATE TABLE url_stats_daily
(
    id          UUID PRIMARY KEY,
    url_id      UUID   NOT NULL REFERENCES urls (id),
    date        DATE   NOT NULL,
    click_count BIGINT NOT NULL DEFAULT 0,

    -- One row per URL per day. This constraint also creates an implicit index
    -- that Postgres will use for lookups by (url_id, date).
    CONSTRAINT url_stats_daily_url_id_date_unique UNIQUE (url_id, date)
);

COMMIT;