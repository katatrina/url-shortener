package link

import "time"

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

type ClickStatsQuery struct {
	LinkID       string
	From         time.Time
	To           time.Time
	PreviousFrom time.Time
	PreviousTo   time.Time
	Bucket       string
	Timezone     string
	TopN         int
}

const (
	BucketHour = "hour"
	BucketDay  = "day"
)

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

type ClickStats struct {
	Summary      ClickSummary
	Timeseries   []TimePoint
	TopCountries []DimensionCount
	TopReferrers []DimensionCount
}
