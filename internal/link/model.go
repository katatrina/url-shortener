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
	LinkID   string
	From     time.Time
	Bucket   string
	Timezone string
	TopN     int
}

// Bucket fields accepted by ClickStatsQuery. They map 1:1 to date_trunc fields.
const (
	BucketHour = "hour"
	BucketDay  = "day"
)

type ClickSummary struct {
	TotalClicks   int64 `db:"total_clicks"`
	ClicksInRange int64 `db:"clicks_in_range"`
}

type TimePoint struct {
	Bucket time.Time `db:"bucket"`
	Clicks int64     `db:"clicks"`
}

// DimensionCount is one row of a top-N breakdown. Value is empty when the
// dimension is unknown: no GeoIP hit for a country, no referrer header for a
// referrer host (i.e. direct traffic). Naming that is the client's job.
type DimensionCount struct {
	Value  string `db:"value"`
	Clicks int64  `db:"clicks"`
}

type ClickStats struct {
	Summary      ClickSummary
	Timeseries   []TimePoint
	TopCountries []DimensionCount
	TopReferrers []DimensionCount
}
