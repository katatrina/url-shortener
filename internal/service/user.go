package service

import (
	"context"

	"github.com/katatrina/url-shortener/internal/model"
)

func (s *Service) GetUserProfile(ctx context.Context, id string) (*model.User, error) {
	return s.userRepo.FindByID(ctx, id)
}
