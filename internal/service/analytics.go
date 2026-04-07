package service

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/katatrina/url-shortener/internal/model"
)

const (
	defaultTopLimit = 10
)

// GetURLStats fetches all analytics data for a URL's dashboard.
//
// The 6 breakdown queries (daily, referrers, countries, OS, browsers, devices)
// are independent — they don't read each other's results — so we run them in
// parallel with errgroup to reduce wall-clock time from ~6x to ~1x.
//
// Why no DB transaction?
// A transaction would force all queries onto a single Postgres connection,
// serializing them and negating the parallelism benefit. The data is already
// approximate by nature (aggregator lag, dropped events), so minor
// inconsistency between breakdown dimensions (a few events difference) is
// acceptable for a dashboard.
func (s *Service) GetURLStats(ctx context.Context, params model.GetURLStatsParams) (*model.URLStatsResponse, error) {
	u, err := s.urlRepo.FindByShortCode(ctx, params.ShortCode)
	if err != nil {
		return nil, err
	}

	if u.UserID == nil || *u.UserID != params.UserID {
		return nil, model.ErrURLOwnerMismatch
	}

	to := time.Now().UTC().Truncate(24 * time.Hour)
	from := to.AddDate(0, 0, -params.Days)

	var (
		dailyStats   []model.DailyStat
		topReferrers []model.ReferrerStat
		topCountries []model.CountryStat
		topOSes      []model.OSStat
		topBrowsers  []model.BrowserStat
		topDevices   []model.DeviceStat
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		dailyStats, err = s.statsRepo.GetDailyStats(gCtx, u.ID, from, to)
		return err
	})

	g.Go(func() error {
		var err error
		topReferrers, err = s.clickEventRepo.GetTopReferrers(gCtx, u.ID, from, to, defaultTopLimit)
		return err
	})

	g.Go(func() error {
		var err error
		topCountries, err = s.clickEventRepo.GetTopCountries(gCtx, u.ID, from, to, defaultTopLimit)
		return err
	})

	g.Go(func() error {
		var err error
		topOSes, err = s.clickEventRepo.GetTopOSes(gCtx, u.ID, from, to, defaultTopLimit)
		return err
	})

	g.Go(func() error {
		var err error
		topBrowsers, err = s.clickEventRepo.GetTopBrowsers(gCtx, u.ID, from, to, defaultTopLimit)
		return err
	})

	g.Go(func() error {
		var err error
		topDevices, err = s.clickEventRepo.GetTopDevices(gCtx, u.ID, from, to, defaultTopLimit)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	dailyResponses := make([]model.DailyStatResponse, len(dailyStats))
	for i, ds := range dailyStats {
		dailyResponses[i] = model.DailyStatResponse{
			Date:       ds.Date.Unix(),
			ClickCount: ds.ClickCount,
		}
	}

	return &model.URLStatsResponse{
		TotalClicks:  u.ClickCount,
		DailyClicks:  dailyResponses,
		TopReferrers: topReferrers,
		TopCountries: topCountries,
		TopOSes:      topOSes,
		TopBrowsers:  topBrowsers,
		TopDevices:   topDevices,
	}, nil
}
