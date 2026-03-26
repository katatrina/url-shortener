package service

import (
	"context"
	"time"

	"github.com/katatrina/url-shortener/internal/analytics"
	"github.com/katatrina/url-shortener/internal/cache"
	"github.com/katatrina/url-shortener/internal/model"
	"github.com/katatrina/url-shortener/internal/token"
)

type URLRepository interface {
	Create(ctx context.Context, url model.URL) (*model.URL, error)
	FindByShortCode(ctx context.Context, shortCode string) (*model.URL, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]model.URL, error)
	CountByUserID(ctx context.Context, userID string) (int64, error)
	Delete(ctx context.Context, id string) error
	ShortCodeExists(ctx context.Context, shortCode string) (bool, error)
}

type UserRepository interface {
	Create(ctx context.Context, user model.User) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}

type URLCacheRepository interface {
	Get(ctx context.Context, shortCode string) (*cache.CachedURL, error)
	Set(ctx context.Context, shortCode string, cached *cache.CachedURL) error
	Delete(ctx context.Context, shortCode string) error
}

// ClickEventQueryRepository defines read operations on click events.
// This is separate from the collector's ClickEventRepository (which only writes)
// because the readers and writers have different consumers:
// - Write: analytics collector (internal, high-throughput)
// - Read: analytics API (user-facing, query-optimized)
type ClickEventQueryRepository interface {
	GetTopReferrers(ctx context.Context, urlID string, limit int) ([]model.ReferrerStat, error)
	GetTopCountries(ctx context.Context, urlID string, limit int) ([]model.CountryStat, error)
}

// URLStatsQueryRepository defines read operations on pre-aggregated stats.
type URLStatsQueryRepository interface {
	GetDailyStats(ctx context.Context, urlID string, from, to time.Time) ([]model.DailyStat, error)
}

type ClickCollector interface {
	Track(event analytics.ClickEvent)
}

type Service struct {
	urlRepo        URLRepository
	userRepo       UserRepository
	urlCache       URLCacheRepository
	clickEventRepo ClickEventQueryRepository
	statsRepo      URLStatsQueryRepository
	tokenMaker     token.TokenMaker
	collector      ClickCollector
}

func New(
	urlRepo URLRepository,
	userRepo UserRepository,
	urlCache URLCacheRepository,
	clickEventRepo ClickEventQueryRepository,
	statsRepo URLStatsQueryRepository,
	tokenMaker token.TokenMaker,
	collector ClickCollector,
) *Service {
	return &Service{
		urlRepo:        urlRepo,
		userRepo:       userRepo,
		urlCache:       urlCache,
		clickEventRepo: clickEventRepo,
		statsRepo:      statsRepo,
		tokenMaker:     tokenMaker,
		collector:      collector,
	}
}
