package link

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/katatrina/url-shortener/internal/slug"
)

const maxSlugRetries = 3

type Service struct {
	linkRepo *Repository
}

func NewService(linkRepo *Repository) *Service {
	return &Service{linkRepo: linkRepo}
}

type CreateLinkParams struct {
	OwnerID        string
	DestinationURL string
	Title          *string
}

func (s *Service) CreateLink(ctx context.Context, arg CreateLinkParams) (*Link, error) {
	for range maxSlugRetries {
		generatedSlug := slug.Generate()
		id, _ := uuid.NewV7()

		link, err := s.linkRepo.Create(ctx, InsertLinkParams{
			ID:             id.String(),
			OwnerID:        arg.OwnerID,
			Slug:           generatedSlug,
			DestinationURL: arg.DestinationURL,
			Title:          arg.Title,
			IsCustomSlug:   false,
		})
		if err != nil {
			if errors.Is(err, ErrSlugExists) {
				continue
			}
			return nil, err
		}

		return link, nil
	}

	return nil, fmt.Errorf("failed to generate a unique slug after %d retries", maxSlugRetries)
}
