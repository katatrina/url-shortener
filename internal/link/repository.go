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
	Title          string
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

func (r *Repository) FindByIDAndUserID(ctx context.Context, id, userID string) (*Link, error) {
	query := `
		SELECT id, user_id, slug, destination_url, title, is_custom_slug, created_at, updated_at
		FROM links
		WHERE id = $1 AND user_id = $2
	`

	rows, _ := r.db.Query(ctx, query, id, userID)
	link, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Link])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	return &link, nil
}

func (r *Repository) List(ctx context.Context, userID string) ([]LinkListItem, error) {
	query := `
		SELECT l.*, (SELECT count(*) FROM clicks WHERE clicks.link_id = l.id) AS click_count
		FROM links l WHERE l.user_id = $1 ORDER BY l.created_at DESC;
	`

	rows, _ := r.db.Query(ctx, query, userID)
	return pgx.CollectRows(rows, pgx.RowToStructByName[LinkListItem])
}

func (r *Repository) Count(ctx context.Context, userID string) (int64, error) {
	query := `SELECT count(*) FROM links WHERE user_id = $1`

	var count int64
	if err := r.db.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) Update(ctx context.Context, arg UpdateLinkParams) (*Link, error) {
	query := `
		UPDATE links
		SET destination_url = COALESCE($3, destination_url),
		    title           = COALESCE($4, title),
		    updated_at      = now()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, slug, destination_url, title, is_custom_slug, created_at, updated_at
	`

	rows, _ := r.db.Query(ctx, query, arg.ID, arg.UserID, arg.DestinationURL, arg.Title)
	link, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Link])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	return &link, nil
}

func (r *Repository) Delete(ctx context.Context, id, userID string) error {
	query := `DELETE FROM links WHERE id = $1 AND user_id = $2`

	tag, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrLinkNotFound
	}

	return nil
}
