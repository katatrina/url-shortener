package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/katatrina/url-shortener/internal/model"
)

// GetDailyStats returns click counts per day for a URL within a date range.
// Uses url_stats_daily (pre-aggregated) for fast queries.
func (r *URLStatsRepository) GetDailyStats(ctx context.Context, urlID string, from, to time.Time) ([]model.DailyStat, error) {
	query := `
		SELECT date, click_count
		FROM url_stats_daily
		WHERE url_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date ASC
	`

	rows, _ := r.db.Query(ctx, query, urlID, from, to)
	stats, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.DailyStat])
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetTopReferrers returns the most common referrer values for a URL.
// Queries click_events directly (raw data) since referrer isn't pre-aggregated.
func (r *ClickEventRepository) GetTopReferrers(ctx context.Context, urlID string, limit int) ([]model.ReferrerStat, error) {
	query := `
		SELECT COALESCE(NULLIF(referer, ''), 'Direct') AS referer, COUNT(*) AS click_count
		FROM click_events
		WHERE url_id = $1
		GROUP BY referer
		ORDER BY click_count DESC
		LIMIT $2
	`

	rows, _ := r.db.Query(ctx, query, urlID, limit)
	stats, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.ReferrerStat])
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetTopCountries returns the most common countries for a URL's clicks.
// Returns "Unknown" for clicks where geo-IP lookup wasn't available.
func (r *ClickEventRepository) GetTopCountries(ctx context.Context, urlID string, limit int) ([]model.CountryStat, error) {
	query := `
		SELECT COALESCE(NULLIF(country, ''), 'Unknown') AS country, COUNT(*) AS click_count
		FROM click_events
		WHERE url_id = $1
		GROUP BY country
		ORDER BY click_count DESC
		LIMIT $2
	`

	rows, _ := r.db.Query(ctx, query, urlID, limit)
	stats, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.CountryStat])
	if err != nil {
		return nil, err
	}

	return stats, nil
}
