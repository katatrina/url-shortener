package link

import (
	"context"

	"github.com/jackc/pgx/v5"
)

const clickStatsSummary = `
	SELECT count(*) FILTER (WHERE clicked_at >= $2 AND clicked_at < $3) AS clicks,
	       count(*) FILTER (WHERE clicked_at >= $4 AND clicked_at < $5) AS previous_clicks
	FROM clicks
	WHERE link_id = $1 AND clicked_at >= $4 AND clicked_at < $3
`

const clickStatsTimeseries = `
	WITH agg AS (
	    SELECT date_trunc($4::text, clicked_at, $5::text) AS bucket,
	           count(*)                                   AS clicks
	    FROM clicks
	    WHERE link_id = $1 AND clicked_at >= $2 AND clicked_at < $3
	    GROUP BY 1
	)
	SELECT series.bucket,
	       COALESCE(agg.clicks, 0) AS clicks
	FROM generate_series($2::timestamptz, $3::timestamptz, ('1 ' || $4::text)::interval, $5::text) AS series(bucket)
	         LEFT JOIN agg ON agg.bucket = series.bucket
	ORDER BY series.bucket
`

const clickStatsTopCountries = `
	SELECT country_code AS value, count(*) AS clicks
	FROM clicks
	WHERE link_id = $1 AND clicked_at >= $2 AND clicked_at < $3
	GROUP BY 1
	ORDER BY clicks DESC, value NULLS LAST
	LIMIT $4
`

const clickStatsTopReferrers = `
	SELECT referrer_host AS value, count(*) AS clicks
	FROM clicks
	WHERE link_id = $1 AND clicked_at >= $2 AND clicked_at < $3
	GROUP BY 1
	ORDER BY clicks DESC, value NULLS LAST
	LIMIT $4
`

func (r *Repository) ClickStats(ctx context.Context, q ClickStatsQuery) (*ClickStats, error) {
	var stats ClickStats

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	err = tx.QueryRow(ctx, clickStatsSummary, q.LinkID, q.From, q.To, q.PreviousFrom, q.PreviousTo).Scan(
		&stats.Summary.Clicks,
		&stats.Summary.PreviousClicks,
	)
	if err != nil {
		return nil, err
	}

	rows, _ := tx.Query(ctx, clickStatsTimeseries, q.LinkID, q.From, q.To, q.Bucket, q.Timezone)
	stats.Timeseries, err = pgx.CollectRows(rows, pgx.RowToStructByName[TimePoint])
	if err != nil {
		return nil, err
	}

	rows, _ = tx.Query(ctx, clickStatsTopCountries, q.LinkID, q.From, q.To, q.TopN)
	stats.TopCountries, err = pgx.CollectRows(rows, pgx.RowToStructByName[DimensionCount])
	if err != nil {
		return nil, err
	}

	rows, _ = tx.Query(ctx, clickStatsTopReferrers, q.LinkID, q.From, q.To, q.TopN)
	stats.TopReferrers, err = pgx.CollectRows(rows, pgx.RowToStructByName[DimensionCount])
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &stats, nil
}
