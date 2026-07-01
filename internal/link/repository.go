package link

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

type InsertLinkParams struct {
	ID             string
	OwnerID        string
	Slug           string
	DestinationURL string
	Title          *string
	IsCustomSlug   bool
}

func (r *Repository) Create(ctx context.Context, arg InsertLinkParams) (*Link, error) {
	query := `
		INSERT INTO links (id, owner_id, slug, destination_url, title, is_custom_slug)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, owner_id, slug, destination_url, title, is_custom_slug, created_at, updated_at
	`

	rows, _ := r.db.Query(ctx, query,
		arg.ID, arg.OwnerID, arg.Slug, arg.DestinationURL, arg.Title, arg.IsCustomSlug,
	)

	link, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Link])
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "links_slug_key" {
				return nil, ErrSlugExists
			}
		}
		return nil, err
	}

	return &link, nil
}
