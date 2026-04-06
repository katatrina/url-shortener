BEGIN;

CREATE TABLE click_events
(
    id         UUID PRIMARY KEY,
    url_id     UUID        NOT NULL REFERENCES urls (id),

    ip_address INET,
    referer    TEXT,
    user_agent TEXT,
    country    TEXT,

    clicked_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_click_events_url_id_clicked_at
    ON click_events (url_id, clicked_at DESC);

CREATE TABLE url_stats_daily
(
    id          UUID PRIMARY KEY,
    url_id      UUID   NOT NULL REFERENCES urls (id),

    date        DATE   NOT NULL,
    click_count BIGINT NOT NULL DEFAULT 0,

    CONSTRAINT url_stats_daily_url_id_date_unique UNIQUE (url_id, date)
);

COMMIT;