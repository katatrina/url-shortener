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

type CreateUserParams struct {
	ID           string
	Email        string
	PasswordHash string
}

func (r *Repository) Create(ctx context.Context, arg CreateUserParams) (*User, error) {
	query := `
		INSERT INTO users (id, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, full_name, password_hash, created_at, updated_at
	`

	rows, _ := r.db.Query(ctx, query,
		arg.ID, arg.Email, arg.PasswordHash,
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

func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, full_name, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	rows, _ := r.db.Query(ctx, query, email)
	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
