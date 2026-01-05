package models

import (
	"database/sql"
	"encoding/json"
)

// Entity represents a data entity within a project
type Entity struct {
	BaseEntity
	ProjectID      int64           `db:"project_id" json:"-"` // FK to projects.id (internal)
	Name           string          `db:"name" json:"name"`
	TableName      string          `db:"table_name" json:"table_name"`
	Description    sql.NullString  `db:"description" json:"-"`
	Fields         json.RawMessage `db:"fields" json:"fields"`           // JSON array of fields
	EndpointsCount int             `db:"endpoints_count" json:"-"`       // Count of endpoints (populated via query)
}

// MarshalJSON custom JSON marshaling for Entity
func (e Entity) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		BaseEntityJSON
		Name           string          `json:"name"`
		TableName      string          `json:"table_name"`
		Description    string          `json:"description,omitempty"`
		Fields         json.RawMessage `json:"fields"`
		EndpointsCount int             `json:"endpoints_count"`
	}{
		BaseEntityJSON: e.BaseEntity.ToJSON(),
		Name:           e.Name,
		TableName:      e.TableName,
		Description:    e.Description.String,
		Fields:         e.Fields,
		EndpointsCount: e.EndpointsCount,
	})
}

// EntityField represents a field in an entity
type EntityField struct {
	Name         string `json:"name"`
	Type         string `json:"type"` // string, int, float, bool, date, datetime, json, relation
	Required     bool   `json:"required"`
	Unique       bool   `json:"unique"`
	DefaultValue string `json:"default_value,omitempty"`
	Length       int    `json:"length,omitempty"` // for string types
	Description  string `json:"description,omitempty"`

	// Relation-specific fields (only when Type = "relation")
	RelationType  string `json:"relation_type,omitempty"`  // belongsTo, hasOne, hasMany, manyToMany
	RelatedEntity string `json:"related_entity,omitempty"` // Target entity name
	ForeignKey    string `json:"foreign_key,omitempty"`    // FK column name (auto-generated if empty)
	OnDelete      string `json:"on_delete,omitempty"`      // CASCADE, SET NULL, RESTRICT (default: CASCADE)
}

// IsRelation checks if this field is a relation type
func (f EntityField) IsRelation() bool {
	return f.Type == "relation"
}

// GetForeignKeyColumn returns the FK column name, auto-generating if not specified
func (f EntityField) GetForeignKeyColumn() string {
	if f.ForeignKey != "" {
		return f.ForeignKey
	}
	// Auto-generate: related_entity_id (e.g., user_id)
	return toSnakeCase(f.RelatedEntity) + "_id"
}

// toSnakeCase converts string to snake_case
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// EntityWithEndpoints includes related endpoints
type EntityWithEndpoints struct {
	Entity
	Endpoints []Endpoint `json:"endpoints,omitempty"`
}

// CreateEntityRequest for creating a new entity
type CreateEntityRequest struct {
	ProjectUUID       string             `json:"project_id"` // Will be set from URL param, not from request body
	Name              string             `json:"name" binding:"required,min=2,max=100"`
	TableName         string             `json:"table_name" binding:"required,min=2,max=100"`
	Description       string             `json:"description" binding:"max=500"`
	Fields            []EntityField      `json:"fields" binding:"required,min=1"`
	GenerateEndpoints *GenerateEndpoints `json:"generate_endpoints,omitempty"`
}

// GenerateEndpoints options for auto-generating CRUD endpoints
type GenerateEndpoints struct {
	List   bool `json:"list"`   // GET /entities
	Get    bool `json:"get"`    // GET /entities/:id
	Create bool `json:"create"` // POST /entities
	Update bool `json:"update"` // PUT /entities/:id
	Delete bool `json:"delete"` // DELETE /entities/:id
}

// UpdateEntityRequest for updating entity
type UpdateEntityRequest struct {
	Name        string        `json:"name" binding:"omitempty,min=2,max=100"`
	TableName   string        `json:"table_name" binding:"omitempty,min=2,max=100"`
	Description string        `json:"description" binding:"max=500"`
	Fields      []EntityField `json:"fields" binding:"omitempty,min=1"`
}
