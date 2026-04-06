package model

import "time"

type User struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	FullName     string    `db:"full_name"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
