package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type URLStatsRepository struct {
	db *pgxpool.Pool
}

func NewURLStatsRepository(db *pgxpool.Pool) *URLStatsRepository {
	return &URLStatsRepository{db}
}

// AggregateDaily counts clicks from click_events for the given date
// and upserts the results into url_stats_daily.
//
// How it works:
// 1. Sub-query groups click_events by (url_id, date), counting clicks per group
// 2. INSERT each group into url_stats_daily
// 3. ON CONFLICT (same url_id + date already exists) → add new count to existing count
//
// Why "add to existing" instead of "replace"?
// Because the job might run multiple times per day. First run at 14:00 counts
// 500 clicks. Second run at 18:00 counts 300 new clicks. We want the daily
// total to be 800, not 300. So we accumulate.
//
// Wait — doesn't that double-count? No, because we only count clicks that
// happened AFTER the latest aggregation. See the clicked_at filter below.
//
// Actually, for simplicity in this version, we use a different strategy:
// we REPLACE the count entirely by re-counting all clicks for that date.
// This is idempotent — running the job 1 time or 10 times gives the same result.
// The trade-off is slightly more work per run, but correctness is guaranteed
// without needing to track "last aggregated at" timestamps.
func (r *URLStatsRepository) AggregateDaily(ctx context.Context, date time.Time) (int64, error) {
	query := `
		INSERT INTO url_stats_daily (id, url_id, date, click_count)
		SELECT gen_random_uuid(), ce.url_id, $1::date, COUNT(*)
		FROM click_events ce
		WHERE ce.clicked_at >= $1::date
		  AND ce.clicked_at < ($1::date + INTERVAL '1 day')
		GROUP BY ce.url_id
		ON CONFLICT (url_id, date)
		DO UPDATE SET click_count = EXCLUDED.click_count
	`

	result, err := r.db.Exec(ctx, query, date)
	if err != nil {
		return 0, fmt.Errorf("aggregate daily failed: %w", err)
	}

	return result.RowsAffected(), nil
}
