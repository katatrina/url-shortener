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
	UserID         string
	Slug           string
	DestinationURL string
	Title          *string
	IsCustomSlug   bool
}

func (r *Repository) Insert(ctx context.Context, arg InsertLinkParams) (*Link, error) {
	query := `
		INSERT INTO links (id, user_id, slug, destination_url, title, is_custom_slug)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, slug, destination_url, title, is_custom_slug, created_at, updated_at
	`

	rows, _ := r.db.Query(ctx, query,
		arg.ID, arg.UserID, arg.Slug, arg.DestinationURL, arg.Title, arg.IsCustomSlug,
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

func (r *Repository) FindBySlug(ctx context.Context, slug string) (*Link, error) {
	query := `
		SELECT id, user_id, slug, destination_url, title, is_custom_slug, created_at, updated_at
		FROM links
		WHERE slug = $1
	`

	rows, _ := r.db.Query(ctx, query, slug)
	link, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Link])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	return &link, nil
}
