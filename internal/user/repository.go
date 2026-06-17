package user

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

func (r *Repository) Create(ctx context.Context, user User) (*User, error) {
	query := `
		INSERT INTO users (id, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, full_name, password_hash, created_at, updated_at
	`

	rows, _ := r.db.Query(ctx, query,
		user.ID, user.Email, user.PasswordHash,
		user.CreatedAt, user.UpdatedAt,
	)

	created, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
				return nil, ErrEmailAlreadyExists
			}
		}
		return nil, err
	}

	return &created, nil
}
