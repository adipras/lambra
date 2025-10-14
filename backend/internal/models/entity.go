package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Entity represents a data entity within a project
type Entity struct {
	ID          int64           `db:"id" json:"id"`
	ProjectID   int64           `db:"project_id" json:"project_id"`
	Name        string          `db:"name" json:"name"`
	TableName   string          `db:"table_name" json:"table_name"`
	Description sql.NullString  `db:"description" json:"description,omitempty"`
	Fields      json.RawMessage `db:"fields" json:"fields"` // JSON array of fields
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
}

// EntityField represents a field in an entity
type EntityField struct {
	Name         string `json:"name"`
	Type         string `json:"type"`         // string, int, float, bool, date, datetime, json
	Required     bool   `json:"required"`
	Unique       bool   `json:"unique"`
	DefaultValue string `json:"default_value,omitempty"`
	Length       int    `json:"length,omitempty"`       // for string types
	Description  string `json:"description,omitempty"`
}

// EntityWithEndpoints includes related endpoints
type EntityWithEndpoints struct {
	Entity
	Endpoints []Endpoint `json:"endpoints,omitempty"`
}

// CreateEntityRequest for creating a new entity
type CreateEntityRequest struct {
	ProjectID   int64         `json:"project_id" binding:"required"`
	Name        string        `json:"name" binding:"required,min=2,max=100"`
	TableName   string        `json:"table_name" binding:"required,min=2,max=100"`
	Description string        `json:"description" binding:"max=500"`
	Fields      []EntityField `json:"fields" binding:"required,min=1"`
}

// UpdateEntityRequest for updating entity
type UpdateEntityRequest struct {
	Name        string        `json:"name" binding:"omitempty,min=2,max=100"`
	TableName   string        `json:"table_name" binding:"omitempty,min=2,max=100"`
	Description string        `json:"description" binding:"max=500"`
	Fields      []EntityField `json:"fields" binding:"omitempty,min=1"`
}
