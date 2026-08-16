package click

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StatsQuery describes one analytics read for a single link.
//
// From is already aligned to a bucket boundary in Timezone by the caller, so
// the generated series and the click filter agree on where the window starts.
// Bucket is the date_trunc field ("hour" or "day") — it reaches SQL as a bind
// parameter, never as concatenated text.
type StatsQuery struct {
	LinkID   string
	From     time.Time
	Bucket   string
	Timezone string
	TopN     int
}

type Summary struct {
	TotalClicks   int64      `db:"total_clicks"`
	ClicksInRange int64      `db:"clicks_in_range"`
	LastClickedAt *time.Time `db:"last_clicked_at"`
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

type Stats struct {
	Summary      Summary
	Timeseries   []TimePoint
	TopCountries []DimensionCount
	TopReferrers []DimensionCount
}

// Bucket fields accepted by StatsQuery. They map 1:1 to date_trunc fields.
const (
	BucketHour = "hour"
	BucketDay  = "day"
)

const statsSummary = `
	SELECT count(*)                                 AS total_clicks,
	       count(*) FILTER (WHERE clicked_at >= $2) AS clicks_in_range,
	       max(clicked_at)                          AS last_clicked_at
	FROM clicks
	WHERE link_id = $1
`

// statsTimeseries returns one row per bucket across the whole window, including
// buckets with zero clicks — filling gaps in SQL keeps the client from having to
// reconstruct the x-axis.
//
// clicked_at is timestamptz (an absolute instant); AT TIME ZONE converts it to
// wall-clock time in the caller's zone so "a day" means their day, not UTC's.
// The final AT TIME ZONE converts each bucket back to an absolute instant for
// the wire.
const statsTimeseries = `
	WITH series AS (
	    SELECT generate_series(
	        $2::timestamptz AT TIME ZONE $4::text,
	        now() AT TIME ZONE $4::text,
	        ('1 ' || $3::text)::interval
	    ) AS bucket
	),
	agg AS (
	    SELECT date_trunc($3::text, clicked_at AT TIME ZONE $4::text) AS bucket,
	           count(*)                                               AS clicks
	    FROM clicks
	    WHERE link_id = $1 AND clicked_at >= $2
	    GROUP BY 1
	)
	SELECT (s.bucket AT TIME ZONE $4::text) AS bucket,
	       COALESCE(a.clicks, 0)            AS clicks
	FROM series s
	         LEFT JOIN agg a ON a.bucket = s.bucket
	ORDER BY s.bucket
`

const statsTopCountries = `
	SELECT COALESCE(country_code, '') AS value, count(*) AS clicks
	FROM clicks
	WHERE link_id = $1 AND clicked_at >= $2
	GROUP BY 1
	ORDER BY clicks DESC, value
	LIMIT $3
`

const statsTopReferrers = `
	SELECT COALESCE(referrer_host, '') AS value, count(*) AS clicks
	FROM clicks
	WHERE link_id = $1 AND clicked_at >= $2
	GROUP BY 1
	ORDER BY clicks DESC, value
	LIMIT $3
`

type StatsRepository struct {
	db *pgxpool.Pool
}

func NewStatsRepository(db *pgxpool.Pool) *StatsRepository {
	return &StatsRepository{db: db}
}

// LinkStats runs the four analytics queries for one link.
//
// They run sequentially: four round trips against an indexed table on a link
// the caller already owns. If this ever shows up in latency, the fix is
// pgx.Batch (one round trip), not four goroutines sharing a pool.
//
// LinkStats does not check ownership — the caller must have done that already.
func (r *StatsRepository) LinkStats(ctx context.Context, q StatsQuery) (*Stats, error) {
	var stats Stats

	err := r.db.QueryRow(ctx, statsSummary, q.LinkID, q.From).Scan(
		&stats.Summary.TotalClicks,
		&stats.Summary.ClicksInRange,
		&stats.Summary.LastClickedAt,
	)
	if err != nil {
		return nil, err
	}

	rows, _ := r.db.Query(ctx, statsTimeseries, q.LinkID, q.From, q.Bucket, q.Timezone)
	stats.Timeseries, err = pgx.CollectRows(rows, pgx.RowToStructByName[TimePoint])
	if err != nil {
		return nil, err
	}

	rows, _ = r.db.Query(ctx, statsTopCountries, q.LinkID, q.From, q.TopN)
	stats.TopCountries, err = pgx.CollectRows(rows, pgx.RowToStructByName[DimensionCount])
	if err != nil {
		return nil, err
	}

	rows, _ = r.db.Query(ctx, statsTopReferrers, q.LinkID, q.From, q.TopN)
	stats.TopReferrers, err = pgx.CollectRows(rows, pgx.RowToStructByName[DimensionCount])
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
