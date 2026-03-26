BEGIN;

-- Stores every individual click on a shortened URL.
CREATE TABLE click_events
(
    id         UUID PRIMARY KEY,
    url_id     UUID        NOT NULL REFERENCES urls (id),

    -- IPv4/IPv6 address of the client. Uses Postgres INET type.
    ip_address INET,

    -- Raw HTTP User-Agent header.
    user_agent TEXT,

    -- Raw HTTP Referer header.
    referer    TEXT,

    -- Country code resolved from ip_address via GeoIP lookup.
    country    TEXT,

    -- Timestamp of the click, set by the application at request time.
    clicked_at TIMESTAMPTZ NOT NULL
);

-- Composite index for the primary query pattern: clicks for a given URL in a time range.
CREATE INDEX idx_click_events_url_id_clicked_at
    ON click_events (url_id, clicked_at DESC);

-- Pre-aggregated daily click counts per URL.
-- Populated by a periodic aggregation job from click_events.
CREATE TABLE url_stats_daily
(
    id          UUID PRIMARY KEY,
    url_id      UUID   NOT NULL REFERENCES urls (id),
    date        DATE   NOT NULL,
    click_count BIGINT NOT NULL DEFAULT 0,

    -- One row per URL per day. Also serves as an implicit index on (url_id, date).
    CONSTRAINT url_stats_daily_url_id_date_unique UNIQUE (url_id, date)
);

COMMIT;