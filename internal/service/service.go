package service

import (
	"context"

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

// URLCacheRepository defines what the service needs from cache.
// The service doesn't know (or care) if this is Redis, Memcached, or a map.
type URLCacheRepository interface {
	Get(ctx context.Context, shortCode string) (*cache.CachedURL, error)
	Set(ctx context.Context, shortCode string, cached *cache.CachedURL) error
	Delete(ctx context.Context, shortCode string) error
}

type Service struct {
	urlRepo    URLRepository
	userRepo   UserRepository
	urlCache   URLCacheRepository // nil = cache disabled
	tokenMaker token.TokenMaker
}

func New(
	urlRepo URLRepository,
	userRepo UserRepository,
	urlCache URLCacheRepository,
	tokenMaker token.TokenMaker,
) *Service {
	return &Service{
		urlRepo,
		userRepo,
		urlCache,
		tokenMaker,
	}
}
