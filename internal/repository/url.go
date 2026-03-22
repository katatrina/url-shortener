package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/katatrina/url-shortener/internal/model"
)

func (r *URLRepository) Create(ctx context.Context, url model.URL) (*model.URL, error) {
	query := `
		INSERT INTO urls (id, short_code, original_url, user_id, click_count, expires_at, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, short_code, original_url, user_id, click_count, expires_at, created_at, updated_at, deleted_at
	`

	rows, _ := r.db.Query(ctx, query,
		url.ID, url.ShortCode, url.OriginalURL, url.UserID,
		url.ClickCount, url.ExpiresAt, url.CreatedAt, url.UpdatedAt, url.DeletedAt,
	)

	created, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.URL])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "urls_short_code_unique" {
				return nil, model.ErrShortCodeTaken
			}
		}
		return nil, err
	}

	return &created, nil
}

// FindByShortCode looks up a URL by its short code.
func (r *URLRepository) FindByShortCode(ctx context.Context, shortCode string) (*model.URL, error) {
	query := `
		SELECT id, short_code, original_url, user_id, click_count, expires_at, created_at, updated_at, deleted_at
		FROM urls
		WHERE short_code = $1 AND deleted_at IS NULL
	`

	rows, _ := r.db.Query(ctx, query, shortCode)
	url, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.URL])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrURLNotFound
		}
		return nil, err
	}

	return &url, nil
}

// IncrementClickCount atomically increments the click counter.
func (r *URLRepository) IncrementClickCount(ctx context.Context, id string) error {
	query := `
		UPDATE urls
		SET click_count = click_count + 1, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return model.ErrURLNotFound
	}

	return nil
}

func (r *URLRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]model.URL, error) {
	query := `
		SELECT id, short_code, original_url, user_id, click_count, expires_at, created_at, updated_at, deleted_at
		FROM urls
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, _ := r.db.Query(ctx, query, userID, limit, offset)
	urls, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.URL])
	if err != nil {
		return nil, err
	}

	return urls, nil
}

func (r *URLRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM urls
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *URLRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE urls
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return model.ErrURLNotFound
	}

	return nil
}

// ShortCodeExists checks if a short code is already in use.
func (r *URLRepository) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1 AND deleted_at IS NULL)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, shortCode).Scan(&exists)
	return exists, err
}
