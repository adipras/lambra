package generator

// Model template
const modelTemplate = `package {{ .PackageName }}

import (
{{- range .Imports }}
	"{{ . }}"
{{- end }}
)

// {{ .EntityName }} represents a {{ toLower .EntityName }} entity
type {{ .EntityName }} struct {
	BaseEntity
{{- range .Fields }}
	{{ .Name }} {{ .GoType }} ` + "`" + `{{ .JSONTag }} {{ .DBTag }}{{ if .ValidateTag }} {{ .ValidateTag }}{{ end }}` + "`" + ` // {{ .Description }}
{{- end }}
}

// TableName returns the table name for {{ .EntityName }}
func ({{ .EntityNameLC }} *{{ .EntityName }}) TableName() string {
	return "{{ .TableName }}"
}

// Validate validates the {{ .EntityName }} model
func ({{ .EntityNameLC }} *{{ .EntityName }}) Validate() error {
{{- range .Fields }}
{{- if .Required }}
	if {{ $.EntityNameLC }}.{{ .Name }} == {{ if eq .GoType "string" }}""{{ else }}nil{{ end }} {
		return fmt.Errorf("{{ .NameLC }} is required")
	}
{{- end }}
{{- end }}
	return nil
}
`

// Repository template
const repositoryTemplate = `package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/internal/models"
)

// {{ .EntityName }}Repository handles {{ toLower .EntityName }} data operations
type {{ .EntityName }}Repository struct {
	db *sqlx.DB
}

// New{{ .EntityName }}Repository creates a new {{ .EntityName }} repository
func New{{ .EntityName }}Repository(db *sqlx.DB) *{{ .EntityName }}Repository {
	return &{{ .EntityName }}Repository{db: db}
}

// Create creates a new {{ toLower .EntityName }}
func (r *{{ .EntityName }}Repository) Create(ctx context.Context, {{ .EntityNameLC }} *models.{{ .EntityName }}) error {
	now := time.Now()
	{{ .EntityNameLC }}.UUID = uuid.New()
	{{ .EntityNameLC }}.CreatedAt = now
	{{ .EntityNameLC }}.UpdatedAt = now

	query := ` + "`" + `
		INSERT INTO {{ .TableName }} (
			uuid,
{{- range .Fields }}
			{{ toSnake .Name }},
{{- end }}
			created_at, updated_at
		) VALUES (
			:uuid,
{{- range .Fields }}
			:{{ toSnake .Name }},
{{- end }}
			:created_at, :updated_at
		) RETURNING id
	` + "`" + `

	rows, err := r.db.NamedQueryContext(ctx, query, {{ .EntityNameLC }})
	if err != nil {
		return fmt.Errorf("failed to create {{ toLower .EntityName }}: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&{{ .EntityNameLC }}.ID); err != nil {
			return fmt.Errorf("failed to scan id: %w", err)
		}
	}

	return nil
}

// GetByID retrieves a {{ toLower .EntityName }} by ID
func (r *{{ .EntityName }}Repository) GetByID(ctx context.Context, id int64) (*models.{{ .EntityName }}, error) {
	var {{ .EntityNameLC }} models.{{ .EntityName }}
	query := ` + "`" + `SELECT * FROM {{ .TableName }} WHERE id = $1 AND deleted_at IS NULL` + "`" + `

	if err := r.db.GetContext(ctx, &{{ .EntityNameLC }}, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("{{ toLower .EntityName }} not found")
		}
		return nil, fmt.Errorf("failed to get {{ toLower .EntityName }}: %w", err)
	}

	return &{{ .EntityNameLC }}, nil
}

// GetByUUID retrieves a {{ toLower .EntityName }} by UUID
func (r *{{ .EntityName }}Repository) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.{{ .EntityName }}, error) {
	var {{ .EntityNameLC }} models.{{ .EntityName }}
	query := ` + "`" + `SELECT * FROM {{ .TableName }} WHERE uuid = $1 AND deleted_at IS NULL` + "`" + `

	if err := r.db.GetContext(ctx, &{{ .EntityNameLC }}, query, uuid); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("{{ toLower .EntityName }} not found")
		}
		return nil, fmt.Errorf("failed to get {{ toLower .EntityName }}: %w", err)
	}

	return &{{ .EntityNameLC }}, nil
}

// List retrieves all {{ pluralize (toLower .EntityName) }}
func (r *{{ .EntityName }}Repository) List(ctx context.Context, limit, offset int) ([]*models.{{ .EntityName }}, error) {
	var {{ .EntityNameLC }}s []*models.{{ .EntityName }}
	query := ` + "`" + `
		SELECT * FROM {{ .TableName }}
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	` + "`" + `

	if err := r.db.SelectContext(ctx, &{{ .EntityNameLC }}s, query, limit, offset); err != nil {
		return nil, fmt.Errorf("failed to list {{ pluralize (toLower .EntityName) }}: %w", err)
	}

	return {{ .EntityNameLC }}s, nil
}

// Update updates a {{ toLower .EntityName }}
func (r *{{ .EntityName }}Repository) Update(ctx context.Context, {{ .EntityNameLC }} *models.{{ .EntityName }}) error {
	{{ .EntityNameLC }}.UpdatedAt = time.Now()

	query := ` + "`" + `
		UPDATE {{ .TableName }} SET
{{- range $i, $field := .Fields }}
			{{ toSnake .Name }} = :{{ toSnake .Name }},
{{- end }}
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	` + "`" + `

	result, err := r.db.NamedExecContext(ctx, query, {{ .EntityNameLC }})
	if err != nil {
		return fmt.Errorf("failed to update {{ toLower .EntityName }}: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("{{ toLower .EntityName }} not found")
	}

	return nil
}

// Delete soft deletes a {{ toLower .EntityName }}
func (r *{{ .EntityName }}Repository) Delete(ctx context.Context, id int64) error {
	query := ` + "`" + `UPDATE {{ .TableName }} SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL` + "`" + `

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete {{ toLower .EntityName }}: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("{{ toLower .EntityName }} not found")
	}

	return nil
}

// Count returns the total count of {{ pluralize (toLower .EntityName) }}
func (r *{{ .EntityName }}Repository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := ` + "`" + `SELECT COUNT(*) FROM {{ .TableName }} WHERE deleted_at IS NULL` + "`" + `

	if err := r.db.GetContext(ctx, &count, query); err != nil {
		return 0, fmt.Errorf("failed to count {{ pluralize (toLower .EntityName) }}: %w", err)
	}

	return count, nil
}
`

// Service template
const serviceTemplate = `package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
)

// {{ .EntityName }}Service handles business logic for {{ pluralize (toLower .EntityName) }}
type {{ .EntityName }}Service struct {
	repo *repository.{{ .EntityName }}Repository
}

// New{{ .EntityName }}Service creates a new {{ .EntityName }} service
func New{{ .EntityName }}Service(repo *repository.{{ .EntityName }}Repository) *{{ .EntityName }}Service {
	return &{{ .EntityName }}Service{
		repo: repo,
	}
}

// Create creates a new {{ toLower .EntityName }}
func (s *{{ .EntityName }}Service) Create(ctx context.Context, {{ .EntityNameLC }} *models.{{ .EntityName }}) error {
	// Validate
	if err := {{ .EntityNameLC }}.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create
	if err := s.repo.Create(ctx, {{ .EntityNameLC }}); err != nil {
		return fmt.Errorf("failed to create {{ toLower .EntityName }}: %w", err)
	}

	return nil
}

// GetByID retrieves a {{ toLower .EntityName }} by ID
func (s *{{ .EntityName }}Service) GetByID(ctx context.Context, id int64) (*models.{{ .EntityName }}, error) {
	{{ .EntityNameLC }}, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get {{ toLower .EntityName }}: %w", err)
	}
	return {{ .EntityNameLC }}, nil
}

// GetByUUID retrieves a {{ toLower .EntityName }} by UUID
func (s *{{ .EntityName }}Service) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.{{ .EntityName }}, error) {
	{{ .EntityNameLC }}, err := s.repo.GetByUUID(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get {{ toLower .EntityName }}: %w", err)
	}
	return {{ .EntityNameLC }}, nil
}

// List retrieves all {{ pluralize (toLower .EntityName) }}
func (s *{{ .EntityName }}Service) List(ctx context.Context, limit, offset int) ([]*models.{{ .EntityName }}, int64, error) {
	{{ .EntityNameLC }}s, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list {{ pluralize (toLower .EntityName) }}: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count {{ pluralize (toLower .EntityName) }}: %w", err)
	}

	return {{ .EntityNameLC }}s, total, nil
}

// Update updates a {{ toLower .EntityName }}
func (s *{{ .EntityName }}Service) Update(ctx context.Context, id int64, {{ .EntityNameLC }} *models.{{ .EntityName }}) error {
	// Check if exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("{{ toLower .EntityName }} not found: %w", err)
	}

	// Update fields
	{{ .EntityNameLC }}.ID = existing.ID
	{{ .EntityNameLC }}.UUID = existing.UUID
	{{ .EntityNameLC }}.CreatedAt = existing.CreatedAt

	// Validate
	if err := {{ .EntityNameLC }}.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Update
	if err := s.repo.Update(ctx, {{ .EntityNameLC }}); err != nil {
		return fmt.Errorf("failed to update {{ toLower .EntityName }}: %w", err)
	}

	return nil
}

// Delete deletes a {{ toLower .EntityName }}
func (s *{{ .EntityName }}Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete {{ toLower .EntityName }}: %w", err)
	}
	return nil
}
`

// Handler template
const handlerTemplate = `package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/lambra/internal/api/dto"
	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/service"
)

// {{ .EntityName }}Handler handles HTTP requests for {{ pluralize (toLower .EntityName) }}
type {{ .EntityName }}Handler struct {
	service *service.{{ .EntityName }}Service
}

// New{{ .EntityName }}Handler creates a new {{ .EntityName }} handler
func New{{ .EntityName }}Handler(service *service.{{ .EntityName }}Service) *{{ .EntityName }}Handler {
	return &{{ .EntityName }}Handler{
		service: service,
	}
}

// Create{{ .EntityName }} creates a new {{ toLower .EntityName }}
func (h *{{ .EntityName }}Handler) Create{{ .EntityName }}(c *gin.Context) {
	var req dto.Create{{ .EntityName }}Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	{{ .EntityNameLC }} := req.ToModel()
	if err := h.service.Create(c.Request.Context(), {{ .EntityNameLC }}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.{{ .EntityName }}Response{}.FromModel({{ .EntityNameLC }}))
}

// Get{{ .EntityName }} retrieves a {{ toLower .EntityName }} by ID
func (h *{{ .EntityName }}Handler) Get{{ .EntityName }}(c *gin.Context) {
	idStr := c.Param("id")

	// Try as UUID first
	if id, err := uuid.Parse(idStr); err == nil {
		{{ .EntityNameLC }}, err := h.service.GetByUUID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "{{ .EntityName }} not found"})
			return
		}
		c.JSON(http.StatusOK, dto.{{ .EntityName }}Response{}.FromModel({{ .EntityNameLC }}))
		return
	}

	// Try as integer ID
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	{{ .EntityNameLC }}, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "{{ .EntityName }} not found"})
		return
	}

	c.JSON(http.StatusOK, dto.{{ .EntityName }}Response{}.FromModel({{ .EntityNameLC }}))
}

// List{{ pluralize .EntityName }} retrieves all {{ pluralize (toLower .EntityName) }}
func (h *{{ .EntityName }}Handler) List{{ pluralize .EntityName }}(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	{{ .EntityNameLC }}s, total, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.{{ .EntityName }}Response, len({{ .EntityNameLC }}s))
	for i, {{ .EntityNameLC }} := range {{ .EntityNameLC }}s {
		response[i] = dto.{{ .EntityName }}Response{}.FromModel({{ .EntityNameLC }})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}

// Update{{ .EntityName }} updates a {{ toLower .EntityName }}
func (h *{{ .EntityName }}Handler) Update{{ .EntityName }}(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req dto.Update{{ .EntityName }}Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	{{ .EntityNameLC }} := req.ToModel()
	if err := h.service.Update(c.Request.Context(), id, {{ .EntityNameLC }}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.{{ .EntityName }}Response{}.FromModel({{ .EntityNameLC }}))
}

// Delete{{ .EntityName }} deletes a {{ toLower .EntityName }}
func (h *{{ .EntityName }}Handler) Delete{{ .EntityName }}(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
`

// DTO template
const dtoTemplate = `package dto

import (
	"github.com/google/uuid"
	"github.com/yourusername/lambra/internal/models"
	"time"
)

// Create{{ .EntityName }}Request represents a request to create a {{ toLower .EntityName }}
type Create{{ .EntityName }}Request struct {
{{- range .Fields }}
	{{ .Name }} {{ .GoType }} ` + "`" + `{{ .JSONTag }}{{ if .ValidateTag }} {{ .ValidateTag }}{{ end }}` + "`" + `
{{- end }}
}

// ToModel converts the request to a model
func (r Create{{ .EntityName }}Request) ToModel() *models.{{ .EntityName }} {
	return &models.{{ .EntityName }}{
{{- range .Fields }}
		{{ .Name }}: r.{{ .Name }},
{{- end }}
	}
}

// Update{{ .EntityName }}Request represents a request to update a {{ toLower .EntityName }}
type Update{{ .EntityName }}Request struct {
{{- range .Fields }}
	{{ .Name }} {{ .GoType }} ` + "`" + `{{ .JSONTag }}{{ if .ValidateTag }} {{ .ValidateTag }}{{ end }}` + "`" + `
{{- end }}
}

// ToModel converts the request to a model
func (r Update{{ .EntityName }}Request) ToModel() *models.{{ .EntityName }} {
	return &models.{{ .EntityName }}{
{{- range .Fields }}
		{{ .Name }}: r.{{ .Name }},
{{- end }}
	}
}

// {{ .EntityName }}Response represents a {{ toLower .EntityName }} response
type {{ .EntityName }}Response struct {
	ID        int64     ` + "`json:\"id\"`" + `
	UUID      uuid.UUID ` + "`json:\"uuid\"`" + `
{{- range .Fields }}
	{{ .Name }} {{ .GoType }} ` + "`" + `{{ .JSONTag }}` + "`" + `
{{- end }}
	CreatedAt time.Time  ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time  ` + "`json:\"updated_at\"`" + `
	DeletedAt *time.Time ` + "`json:\"deleted_at,omitempty\"`" + `
}

// FromModel converts a model to a response
func (r {{ .EntityName }}Response) FromModel({{ .EntityNameLC }} *models.{{ .EntityName }}) {{ .EntityName }}Response {
	return {{ .EntityName }}Response{
		ID:        {{ .EntityNameLC }}.ID,
		UUID:      {{ .EntityNameLC }}.UUID,
{{- range .Fields }}
		{{ .Name }}: {{ $.EntityNameLC }}.{{ .Name }},
{{- end }}
		CreatedAt: {{ .EntityNameLC }}.CreatedAt,
		UpdatedAt: {{ .EntityNameLC }}.UpdatedAt,
		DeletedAt: {{ .EntityNameLC }}.DeletedAt,
	}
}
`

// Migration up template
const migrationUpTemplate = `-- Create {{ .TableName }} table
CREATE TABLE IF NOT EXISTS {{ .TableName }} (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
{{- range .Fields }}
    {{ toSnake .Name }} {{ .Type }}{{ if .Required }} NOT NULL{{ end }}{{ if .DefaultValue }} DEFAULT {{ .DefaultValue }}{{ end }},
{{- end }}
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_{{ .TableName }}_uuid ON {{ .TableName }}(uuid);
CREATE INDEX idx_{{ .TableName }}_deleted_at ON {{ .TableName }}(deleted_at);
CREATE INDEX idx_{{ .TableName }}_created_at ON {{ .TableName }}(created_at);
`

// Migration down template
const migrationDownTemplate = `-- Drop {{ .TableName }} table
DROP TABLE IF EXISTS {{ .TableName }};
`
