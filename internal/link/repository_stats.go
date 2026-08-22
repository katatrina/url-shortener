package link

import (
	"context"

	"github.com/jackc/pgx/v5"
)

const clickStatsSummary = `
	SELECT count(*)                                 AS total_clicks,
	       count(*) FILTER (WHERE clicked_at >= $2) AS clicks_in_range
	FROM clicks
	WHERE link_id = $1
`

// clickStatsTimeseries returns one row per bucket across the whole window,
// including buckets with zero clicks — filling gaps in SQL keeps the client from
// having to reconstruct the x-axis.
//
// clicked_at is timestamptz (an absolute instant); AT TIME ZONE converts it to
// wall-clock time in the caller's zone so "a day" means their day, not UTC's.
// The final AT TIME ZONE converts each bucket back to an absolute instant for
// the wire.
const clickStatsTimeseries = `
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

const clickStatsTopCountries = `
	SELECT COALESCE(country_code, '') AS value, count(*) AS clicks
	FROM clicks
	WHERE link_id = $1 AND clicked_at >= $2
	GROUP BY 1
	ORDER BY clicks DESC, value
	LIMIT $3
`

const clickStatsTopReferrers = `
	SELECT COALESCE(referrer_host, '') AS value, count(*) AS clicks
	FROM clicks
	WHERE link_id = $1 AND clicked_at >= $2
	GROUP BY 1
	ORDER BY clicks DESC, value
	LIMIT $3
`

func (r *Repository) ClickStats(ctx context.Context, q ClickStatsQuery) (*ClickStats, error) {
	var stats ClickStats

	err := r.db.QueryRow(ctx, clickStatsSummary, q.LinkID, q.From).Scan(
		&stats.Summary.TotalClicks,
		&stats.Summary.ClicksInRange,
	)
	if err != nil {
		return nil, err
	}

	rows, _ := r.db.Query(ctx, clickStatsTimeseries, q.LinkID, q.From, q.Bucket, q.Timezone)
	stats.Timeseries, err = pgx.CollectRows(rows, pgx.RowToStructByName[TimePoint])
	if err != nil {
		return nil, err
	}

	rows, _ = r.db.Query(ctx, clickStatsTopCountries, q.LinkID, q.From, q.TopN)
	stats.TopCountries, err = pgx.CollectRows(rows, pgx.RowToStructByName[DimensionCount])
	if err != nil {
		return nil, err
	}

	rows, _ = r.db.Query(ctx, clickStatsTopReferrers, q.LinkID, q.From, q.TopN)
	stats.TopReferrers, err = pgx.CollectRows(rows, pgx.RowToStructByName[DimensionCount])
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
