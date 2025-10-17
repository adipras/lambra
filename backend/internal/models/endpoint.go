package models

import (
	"database/sql"
	"encoding/json"
)

// Endpoint represents an API endpoint
type Endpoint struct {
	BaseEntity
	EntityID       int64           `db:"entity_id" json:"-"`  // FK to entities.id (internal)
	ProjectID      int64           `db:"project_id" json:"-"` // FK to projects.id (internal)
	Name           string          `db:"name" json:"name"`
	Path           string          `db:"path" json:"path"`
	Method         string          `db:"method" json:"method"` // GET, POST, PUT, DELETE, PATCH
	Description    sql.NullString  `db:"description" json:"-"`
	RequestSchema  json.RawMessage `db:"request_schema" json:"request_schema,omitempty"`
	ResponseSchema json.RawMessage `db:"response_schema" json:"response_schema,omitempty"`
	RequireAuth    bool            `db:"require_auth" json:"require_auth"`
}

// MarshalJSON custom JSON marshaling for Endpoint
func (e Endpoint) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		BaseEntityJSON
		Name           string          `json:"name"`
		Path           string          `json:"path"`
		Method         string          `json:"method"`
		Description    string          `json:"description,omitempty"`
		RequestSchema  json.RawMessage `json:"request_schema,omitempty"`
		ResponseSchema json.RawMessage `json:"response_schema,omitempty"`
		RequireAuth    bool            `json:"require_auth"`
	}{
		BaseEntityJSON: e.BaseEntity.ToJSON(),
		Name:           e.Name,
		Path:           e.Path,
		Method:         e.Method,
		Description:    e.Description.String,
		RequestSchema:  e.RequestSchema,
		ResponseSchema: e.ResponseSchema,
		RequireAuth:    e.RequireAuth,
	})
}

// EndpointWithMetrics includes metrics data
type EndpointWithMetrics struct {
	Endpoint
	TotalRequests   int64   `json:"total_requests"`
	AvgResponseTime float64 `json:"avg_response_time"`
	ErrorRate       float64 `json:"error_rate"`
}

// CreateEndpointRequest for creating endpoint
type CreateEndpointRequest struct {
	EntityUUID     string          `json:"entity_id" binding:"required"` // Accepts entity UUID from frontend
	Name           string          `json:"name" binding:"required,min=2,max=100"`
	Path           string          `json:"path" binding:"required,min=1,max=255"`
	Method         string          `json:"method" binding:"required,oneof=GET POST PUT DELETE PATCH"`
	Description    string          `json:"description" binding:"max=500"`
	RequestSchema  json.RawMessage `json:"request_schema"`
	ResponseSchema json.RawMessage `json:"response_schema"`
	RequireAuth    bool            `json:"require_auth"`
}

// UpdateEndpointRequest for updating endpoint
type UpdateEndpointRequest struct {
	Name           string          `json:"name" binding:"omitempty,min=2,max=100"`
	Path           string          `json:"path" binding:"omitempty,min=1,max=255"`
	Method         string          `json:"method" binding:"omitempty,oneof=GET POST PUT DELETE PATCH"`
	Description    string          `json:"description" binding:"max=500"`
	RequestSchema  json.RawMessage `json:"request_schema"`
	ResponseSchema json.RawMessage `json:"response_schema"`
	RequireAuth    *bool           `json:"require_auth"`
}

// TestEndpointRequest for testing an endpoint
type TestEndpointRequest struct {
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body"`
	Params  map[string]string `json:"params"`
}

// TestEndpointResponse contains test result
type TestEndpointResponse struct {
	StatusCode   int               `json:"status_code"`
	ResponseTime int64             `json:"response_time"` // in milliseconds
	Headers      map[string]string `json:"headers"`
	Body         json.RawMessage   `json:"body"`
	Error        string            `json:"error,omitempty"`
}
