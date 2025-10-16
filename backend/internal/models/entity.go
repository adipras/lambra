package models

import (
	"database/sql"
	"encoding/json"
)

// Entity represents a data entity within a project
type Entity struct {
	BaseEntity
	ProjectID   int64           `db:"project_id" json:"-"` // FK to projects.id (internal)
	Name        string          `db:"name" json:"name"`
	TableName   string          `db:"table_name" json:"table_name"`
	Description sql.NullString  `db:"description" json:"-"`
	Fields      json.RawMessage `db:"fields" json:"fields"` // JSON array of fields
}

// MarshalJSON custom JSON marshaling for Entity
func (e Entity) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		BaseEntityJSON
		Name        string          `json:"name"`
		TableName   string          `json:"table_name"`
		Description string          `json:"description,omitempty"`
		Fields      json.RawMessage `json:"fields"`
	}{
		BaseEntityJSON: e.BaseEntity.ToJSON(),
		Name:           e.Name,
		TableName:      e.TableName,
		Description:    e.Description.String,
		Fields:         e.Fields,
	})
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
	ProjectUUID string        `json:"project_id" binding:"required"` // Accepts project UUID from frontend
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
