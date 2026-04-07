package repository

import (
	"context"
	"fmt"
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

// GetTopReferrers returns the most common referrer values for a URL within a date range.
// Queries click_events directly (raw data) since referrer isn't pre-aggregated.
//
// The date range must match GetDailyStats so the dashboard shows consistent data:
// if daily clicks cover the last 30 days, referrer breakdown should too.
func (r *ClickEventRepository) GetTopReferrers(ctx context.Context, urlID string, from, to time.Time, limit int) ([]model.ReferrerStat, error) {
	query := `
		SELECT COALESCE(NULLIF(referer, ''), 'Direct') AS referer, COUNT(*) AS click_count
		FROM click_events
		WHERE url_id = $1 AND clicked_at >= $2 AND clicked_at < $3
		GROUP BY referer
		ORDER BY click_count DESC
		LIMIT $4
	`

	rows, _ := r.db.Query(ctx, query, urlID, from, to, limit)
	stats, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.ReferrerStat])
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetTopCountries returns the most common countries for a URL's clicks within a date range.
// Returns "Unknown" for clicks where geo-IP lookup wasn't available.
//
// The date range must match GetDailyStats so the dashboard shows consistent data.
func (r *ClickEventRepository) GetTopCountries(ctx context.Context, urlID string, from, to time.Time, limit int) ([]model.CountryStat, error) {
	query := `
		SELECT COALESCE(NULLIF(country, ''), 'Unknown') AS country, COUNT(*) AS click_count
		FROM click_events
		WHERE url_id = $1 AND clicked_at >= $2 AND clicked_at < $3
		GROUP BY country
		ORDER BY click_count DESC
		LIMIT $4
	`

	rows, _ := r.db.Query(ctx, query, urlID, from, to, limit)
	stats, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.CountryStat])
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetTopOSes returns the most common operating systems for a URL's clicks within a date range.
func (r *ClickEventRepository) GetTopOSes(ctx context.Context, urlID string, from, to time.Time, limit int) ([]model.OSStat, error) {
	query := `
		SELECT COALESCE(NULLIF(os, ''), 'Unknown') AS os, COUNT(*) AS click_count
		FROM click_events
		WHERE url_id = $1 AND clicked_at >= $2 AND clicked_at < $3
		GROUP BY os
		ORDER BY click_count DESC
		LIMIT $4
	`

	rows, _ := r.db.Query(ctx, query, urlID, from, to, limit)
	stats, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.OSStat])
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetTopBrowsers returns the most common browsers for a URL's clicks within a date range.
func (r *ClickEventRepository) GetTopBrowsers(ctx context.Context, urlID string, from, to time.Time, limit int) ([]model.BrowserStat, error) {
	query := `
		SELECT COALESCE(NULLIF(browser, ''), 'Unknown') AS browser, COUNT(*) AS click_count
		FROM click_events
		WHERE url_id = $1 AND clicked_at >= $2 AND clicked_at < $3
		GROUP BY browser
		ORDER BY click_count DESC
		LIMIT $4
	`

	rows, _ := r.db.Query(ctx, query, urlID, from, to, limit)
	stats, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.BrowserStat])
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetTopDevices returns the most common device types for a URL's clicks within a date range.
func (r *ClickEventRepository) GetTopDevices(ctx context.Context, urlID string, from, to time.Time, limit int) ([]model.DeviceStat, error) {
	query := `
		SELECT COALESCE(NULLIF(device_type, ''), 'Unknown') AS device_type, COUNT(*) AS click_count
		FROM click_events
		WHERE url_id = $1 AND clicked_at >= $2 AND clicked_at < $3
		GROUP BY device_type
		ORDER BY click_count DESC
		LIMIT $4
	`

	rows, _ := r.db.Query(ctx, query, urlID, from, to, limit)
	stats, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.DeviceStat])
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// SyncClickCounts updates urls.click_count from url_stats_daily totals.
//
// This is a denormalization step: we're copying aggregated data back
// into the urls table so that listing URLs doesn't require a JOIN.
//
// Why not just JOIN on read?
//   - ListByUserID is called frequently (every dashboard page load)
//   - JOIN + SUM across url_stats_daily adds latency per query
//   - Denormalized click_count is a single column read — O(1) vs O(N days)
//
// Trade-off: click_count may lag behind by one aggregation interval.
// For a dashboard display, this is perfectly acceptable.
func (r *URLStatsRepository) SyncClickCounts(ctx context.Context) (int64, error) {
	query := `
		UPDATE urls u
		SET click_count = COALESCE(s.total, 0),
		    updated_at = NOW()
		FROM (
			SELECT url_id, SUM(click_count) AS total
			FROM url_stats_daily
			GROUP BY url_id
		) s
		WHERE u.id = s.url_id
		  AND u.click_count != s.total
	`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("sync click counts failed: %w", err)
	}

	return result.RowsAffected(), nil
}
