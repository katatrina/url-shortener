package link

import "time"

const (
	RangeLast24Hours = "24h"
	RangeLast7Days   = "7d"
	RangeLast30Days  = "30d"
	RangeLast90Days  = "90d"
)

const (
	BucketHour = "hour"
	BucketDay  = "day"
)

type Link struct {
	ID             string    `db:"id"`
	UserID         string    `db:"user_id"`
	Slug           string    `db:"slug"`
	DestinationURL string    `db:"destination_url"`
	Title          string    `db:"title"`
	IsCustomSlug   bool      `db:"is_custom_slug"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type LinkListItem struct {
	Link
	ClickCount int64 `db:"click_count"`
}

type ClickSummary struct {
	Clicks         int64 `db:"clicks"`
	PreviousClicks int64 `db:"previous_clicks"`
}

type TimePoint struct {
	Bucket time.Time `db:"bucket"`
	Clicks int64     `db:"clicks"`
}

type DimensionCount struct {
	Value  *string `db:"value"`
	Clicks int64   `db:"clicks"`
}
