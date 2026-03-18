package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/katatrina/url-shortener/internal/model"
)

func (r *UserRepository) Create(ctx context.Context, user model.User) (*model.User, error) {
	query := `
		INSERT INTO users (id, email, display_name, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, display_name, password_hash, created_at, updated_at
	`

	rows, _ := r.db.Query(ctx, query,
		user.ID, user.Email, user.DisplayName, user.PasswordHash,
		user.CreatedAt, user.UpdatedAt,
	)

	created, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
				return nil, model.ErrEmailAlreadyExists
			}
		}
		return nil, err
	}

	return &created, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, display_name, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	rows, _ := r.db.Query(ctx, query, email)
	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
