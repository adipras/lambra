package models

import (
	"encoding/json"
	"time"
)

// User represents a platform user for authentication
type User struct {
	ID           int64      `db:"id" json:"-"`
	UUID         string     `db:"uuid" json:"id"`
	Username     string     `db:"username" json:"username"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Name         string     `db:"name" json:"name"`
	Role         string     `db:"role" json:"role"`
	CreatedBy    *string    `db:"created_by" json:"-"`
	UpdatedBy    *string    `db:"updated_by" json:"-"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at" json:"-"`
}

// MarshalJSON hides password hash and exposes UUID as id
func (u User) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		UUID      string    `json:"id"`
		Username  string    `json:"username"`
		Name      string    `json:"name"`
		Role      string    `json:"role"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}{
		UUID:      u.UUID,
		Username:  u.Username,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	})
}

// LoginRequest for authentication
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse returned on successful authentication
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
