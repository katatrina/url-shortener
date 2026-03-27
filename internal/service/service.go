package service

import (
	"context"
	"time"

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

type ClickEventQueryRepository interface {
	GetTopReferrers(ctx context.Context, urlID string, limit int) ([]model.ReferrerStat, error)
	GetTopCountries(ctx context.Context, urlID string, limit int) ([]model.CountryStat, error)
}

type URLStatsQueryRepository interface {
	GetDailyStats(ctx context.Context, urlID string, from, to time.Time) ([]model.DailyStat, error)
}

type ClickCollector interface {
	Track(urlID string, meta model.ClickMeta)
}

type Service struct {
	urlRepo        URLRepository
	userRepo       UserRepository
	urlCache       URLCacheRepository
	clickEventRepo ClickEventQueryRepository
	statsRepo      URLStatsQueryRepository
	tokenMaker     token.TokenMaker
	clickCollector ClickCollector
}

func New(
	urlRepo URLRepository,
	userRepo UserRepository,
	urlCache URLCacheRepository,
	clickEventRepo ClickEventQueryRepository,
	statsRepo URLStatsQueryRepository,
	tokenMaker token.TokenMaker,
	clickCollector ClickCollector,
) *Service {
	return &Service{
		urlRepo:        urlRepo,
		userRepo:       userRepo,
		urlCache:       urlCache,
		clickEventRepo: clickEventRepo,
		statsRepo:      statsRepo,
		tokenMaker:     tokenMaker,
		clickCollector: clickCollector,
	}
}
