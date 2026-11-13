package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/internal/models"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByUsername retrieves an active (non-deleted) user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, uuid, username, password_hash, name, role,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM users
		WHERE username = ? AND deleted_at IS NULL
	`
	if err := r.db.Get(&user, query, username); err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// GetByUUID retrieves an active user by UUID
func (r *UserRepository) GetByUUID(uuid string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, uuid, username, password_hash, name, role,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM users
		WHERE uuid = ? AND deleted_at IS NULL
	`
	if err := r.db.Get(&user, query, uuid); err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}
