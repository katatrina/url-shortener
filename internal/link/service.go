package link

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/katatrina/url-shortener/internal/slug"
)

const maxSlugRetries = 3

type Service struct {
	linkRepo        *Repository
	maxLinksPerUser int
}

func NewService(linkRepo *Repository, maxLinksPerUser int) *Service {
	return &Service{
		linkRepo:        linkRepo,
		maxLinksPerUser: maxLinksPerUser,
	}
}

type CreateLinkParams struct {
	UserID         string
	DestinationURL string
	Title          string
	Slug           *string
}

func (s *Service) CreateLink(ctx context.Context, arg CreateLinkParams) (*Link, error) {
	// There is a very little chance of race condition here. But it's fine.
	count, err := s.linkRepo.CountByUserID(ctx, arg.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to count user links: %w", err)
	}
	if count >= int64(s.maxLinksPerUser) {
		return nil, ErrLinkQuotaExceeded
	}

	if arg.Slug != nil {
		return s.createWithCustomSlug(ctx, arg)
	}
	return s.createWithGeneratedSlug(ctx, arg)
}

func (s *Service) createWithGeneratedSlug(ctx context.Context, arg CreateLinkParams) (*Link, error) {
	for n := range maxSlugRetries {
		generatedSlug := slug.Generate()
		id, _ := uuid.NewV7()

		link, err := s.linkRepo.Insert(ctx, InsertLinkParams{
			ID:             id.String(),
			UserID:         arg.UserID,
			Slug:           generatedSlug,
			DestinationURL: arg.DestinationURL,
			Title:          arg.Title,
			IsCustomSlug:   false,
		})
		if err != nil {
			if errors.Is(err, ErrSlugExists) {
				slog.WarnContext(ctx, "slug collision, retrying",
					slog.String("slug", generatedSlug),
					slog.Int("attempt", n+1),
				)
				continue
			}
			return nil, err
		}

		return link, nil
	}

	return nil, fmt.Errorf("failed to generate a unique slug after %d retries", maxSlugRetries)
}

func (s *Service) createWithCustomSlug(ctx context.Context, arg CreateLinkParams) (*Link, error) {
	id, _ := uuid.NewV7()
	return s.linkRepo.Insert(ctx, InsertLinkParams{
		ID:             id.String(),
		UserID:         arg.UserID,
		Slug:           *arg.Slug,
		DestinationURL: arg.DestinationURL,
		Title:          arg.Title,
		IsCustomSlug:   true,
	})
}

func (s *Service) ResolveSlug(ctx context.Context, rawSlug string) (*Link, error) {
	return s.linkRepo.FindBySlug(ctx, rawSlug)
}

func (s *Service) ListLinks(ctx context.Context, userID string) ([]Link, error) {
	return s.linkRepo.ListByUserID(ctx, userID)
}

func (s *Service) DeleteLink(ctx context.Context, id, userID string) error {
	return s.linkRepo.DeleteByIDAndUserID(ctx, id, userID)
}
