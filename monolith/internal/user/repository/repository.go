package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kanchan755/wallet_app/monolith/internal/user/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	CreateTx(ctx context.Context, tx *sql.Tx, u *model.User) error
}

type mysqlUserRepository struct {
	db *sql.DB // Assuming you have a database connection here
}

func NewMySQLUserRepository(db *sql.DB) UserRepository {
	return &mysqlUserRepository{
		db: db,
	}
}

func (r *mysqlUserRepository) Create(ctx context.Context, u *model.User) error {
	query := "INSERT INTO users (id, full_name, email, password_hash) VALUES (?, ?, ?, ?)"
	_, err := r.db.ExecContext(ctx, query, u.ID, u.FullName, u.Email, u.PasswordHash)
	return err
}

func (r *mysqlUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	query := "SELECT id, full_name, email, password_hash, created_at, updated_at, deleted_at FROM users WHERE id = ?"
	row := r.db.QueryRowContext(ctx, query, id)

	var u model.User
	err := row.Scan(&u.ID, &u.FullName, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *mysqlUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := "SELECT id, full_name, email, password_hash, created_at, updated_at, deleted_at FROM users WHERE email = ?"
	row := r.db.QueryRowContext(ctx, query, email)

	var u model.User
	err := row.Scan(&u.ID, &u.FullName, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *mysqlUserRepository) Update(ctx context.Context, u *model.User) error {
	query := "UPDATE users SET full_name = ?, email = ?, updated_at = ? WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, u.FullName, u.Email, u.UpdatedAt, u.ID)
	return err
}

func (r *mysqlUserRepository) CreateTx(ctx context.Context, tx *sql.Tx, u *model.User) error {
	query := "INSERT INTO users (id, full_name, email, password_hash) VALUES (?, ?, ?, ?)"
	_, err := tx.ExecContext(ctx, query, u.ID, u.FullName, u.Email, u.PasswordHash)
	return err
}
