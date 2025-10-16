package models

import (
	"database/sql"
	"time"
)

// BaseEntity contains common fields for all entities
// Uses dual identifier strategy: ID (BIGINT) for internal, UUID (CHAR36) for external
type BaseEntity struct {
	ID        int64          `db:"id" json:"-"`                 // Internal ID (not exposed in API)
	UUID      string         `db:"uuid" json:"id"`              // External UUID (exposed as "id" in API)
	CreatedBy sql.NullString `db:"created_by" json:"-"`
	UpdatedBy sql.NullString `db:"updated_by" json:"-"`
	DeletedBy sql.NullString `db:"deleted_by" json:"-"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at,omitempty"`
}

// BaseEntityJSON for JSON marshaling with proper handling of null values
type BaseEntityJSON struct {
	UUID      string     `json:"id"`                   // UUID exposed as "id"
	CreatedBy string     `json:"created_by,omitempty"`
	UpdatedBy string     `json:"updated_by,omitempty"`
	DeletedBy string     `json:"deleted_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ToJSON converts BaseEntity to BaseEntityJSON with proper null handling
func (b *BaseEntity) ToJSON() BaseEntityJSON {
	result := BaseEntityJSON{
		UUID:      b.UUID,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}

	if b.CreatedBy.Valid {
		result.CreatedBy = b.CreatedBy.String
	}
	if b.UpdatedBy.Valid {
		result.UpdatedBy = b.UpdatedBy.String
	}
	if b.DeletedBy.Valid {
		result.DeletedBy = b.DeletedBy.String
	}
	if b.DeletedAt.Valid {
		result.DeletedAt = &b.DeletedAt.Time
	}

	return result
}

// SetCreatedBy sets the created_by field
func (b *BaseEntity) SetCreatedBy(user string) {
	b.CreatedBy = sql.NullString{String: user, Valid: true}
}

// SetUpdatedBy sets the updated_by field
func (b *BaseEntity) SetUpdatedBy(user string) {
	b.UpdatedBy = sql.NullString{String: user, Valid: true}
}

// SetDeletedBy sets the deleted_by field and deleted_at timestamp
func (b *BaseEntity) SetDeletedBy(user string) {
	b.DeletedBy = sql.NullString{String: user, Valid: true}
	b.DeletedAt = sql.NullTime{Time: time.Now(), Valid: true}
}

// IsDeleted checks if the entity is soft deleted
func (b *BaseEntity) IsDeleted() bool {
	return b.DeletedAt.Valid
}
