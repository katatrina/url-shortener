package service

import (
	"context"
	"time"

	"github.com/katatrina/url-shortener/internal/model"
)

type URLRepository interface {
	Create(ctx context.Context, url model.URL) (*model.URL, error)
	FindByShortCode(ctx context.Context, shortCode string) (*model.URL, error)
	IncrementClickCount(ctx context.Context, id string) error
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]model.URL, error)
	CountByUserID(ctx context.Context, userID string) (int64, error)
	Delete(ctx context.Context, id string) error
	ShortCodeExists(ctx context.Context, shortCode string) (bool, error)
}

type UserRepository interface {
	Create(ctx context.Context, user model.User) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
}

type TokenMaker interface {
	CreateToken(userID string) (string, time.Time, error)
	VerifyToken(tokenStr string) (string, error)
}

type Service struct {
	urlRepo    URLRepository
	userRepo   UserRepository
	tokenMaker TokenMaker
}

func New(
	urlRepo URLRepository,
	userRepo UserRepository,
	tokenMaker TokenMaker,
) *Service {
	return &Service{
		urlRepo,
		userRepo,
		tokenMaker,
	}
}
