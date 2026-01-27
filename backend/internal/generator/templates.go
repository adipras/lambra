package generator

// Model template
const modelTemplate = `package {{ .PackageName }}

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// {{ .EntityName }} represents a {{ toLower .EntityName }} entity
type {{ .EntityName }} struct {
	ID        int64          ` + "`" + `json:"-" db:"id"` + "`" + `
	UUID      uuid.UUID      ` + "`" + `json:"id" db:"uuid"` + "`" + `
{{- range .Fields }}
	{{ .Name }} {{ .GoType }} ` + "`" + `{{ .JSONTag }} {{ .DBTag }}{{ if .ValidateTag }} {{ .ValidateTag }}{{ end }}` + "`" + `{{ if .Description }} // {{ .Description }}{{ end }}
{{- end }}
	CreatedAt time.Time      ` + "`" + `json:"created_at" db:"created_at"` + "`" + `
	UpdatedAt time.Time      ` + "`" + `json:"updated_at" db:"updated_at"` + "`" + `
	DeletedAt sql.NullTime   ` + "`" + `json:"-" db:"deleted_at"` + "`" + `
}

// TableName returns the table name for {{ .EntityName }}
func ({{ .EntityNameLC }} *{{ .EntityName }}) TableName() string {
	return "{{ .TableName }}"
}

// Validate validates the {{ .EntityName }} model
func ({{ .EntityNameLC }} *{{ .EntityName }}) Validate() error {
{{- range .Fields }}
{{- if and .Required (not .IsForeignKey) }}
{{- if eq .GoType "bool" }}
	// bool fields are always valid (false is a valid value)
{{- else }}
	if {{ $.EntityNameLC }}.{{ .Name }} == {{ if eq .GoType "string" }}""{{ else if eq .GoType "int" }}0{{ else if eq .GoType "int64" }}0{{ else if eq .GoType "float64" }}0{{ else if eq .GoType "float32" }}0{{ else }}nil{{ end }} {
		return fmt.Errorf("{{ .NameLC }} is required")
	}
{{- end }}
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
	"{{ .ModuleName }}/models"
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
		)
	` + "`" + `

	result, err := r.db.NamedExecContext(ctx, query, {{ .EntityNameLC }})
	if err != nil {
		return fmt.Errorf("failed to create {{ toLower .EntityName }}: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	{{ .EntityNameLC }}.ID = id

	return nil
}

// GetByID retrieves a {{ toLower .EntityName }} by ID
func (r *{{ .EntityName }}Repository) GetByID(ctx context.Context, id int64) (*models.{{ .EntityName }}, error) {
	var {{ .EntityNameLC }} models.{{ .EntityName }}
	query := ` + "`" + `SELECT * FROM {{ .TableName }} WHERE id = ? AND deleted_at IS NULL` + "`" + `

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
	query := ` + "`" + `SELECT * FROM {{ .TableName }} WHERE uuid = ? AND deleted_at IS NULL` + "`" + `

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
		LIMIT ? OFFSET ?
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
	query := ` + "`" + `UPDATE {{ .TableName }} SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL` + "`" + `

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
	"{{ .ModuleName }}/models"
	"{{ .ModuleName }}/repository"
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
// NOTE: All handlers accept ONLY UUID as "id" - internal int64 IDs are never exposed or accepted
const handlerTemplate = `package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"{{ .ModuleName }}/api/dto"
	"{{ .ModuleName }}/service"
)

// {{ .EntityName }}Handler handles HTTP requests for {{ pluralize (toLower .EntityName) }}
type {{ .EntityName }}Handler struct {
	service *service.{{ .EntityName }}Service
	db      *sqlx.DB // For FK UUID→ID translation
}

// New{{ .EntityName }}Handler creates a new {{ .EntityName }} handler
func New{{ .EntityName }}Handler(service *service.{{ .EntityName }}Service, db *sqlx.DB) *{{ .EntityName }}Handler {
	return &{{ .EntityName }}Handler{
		service: service,
		db:      db,
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
	
	// Translate FK UUIDs to internal IDs
{{- range .Fields }}
{{- if .IsForeignKey }}
	if req.{{ .FKName }} != "" {
		{{ .FKNameLC }}, err := uuid.Parse(req.{{ .FKName }})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid {{ .FKNameLC }} format"})
			return
		}
		var fkID int64
		err = h.db.Get(&fkID, "SELECT id FROM {{ toSnake (trimSuffix .Name "Id") }}s WHERE uuid = ? AND deleted_at IS NULL", {{ .FKNameLC }}.String())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "{{ .FKName }} not found"})
			return
		}
		{{ $.EntityNameLC }}.{{ .Name }} = fkID
	}
{{- end }}
{{- end }}
	
	if err := h.service.Create(c.Request.Context(), {{ .EntityNameLC }}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.{{ .EntityName }}Response{}.FromModel({{ .EntityNameLC }}, h.db))
}

// Get{{ .EntityName }} retrieves a {{ toLower .EntityName }} by UUID (query param: id)
func (h *{{ .EntityName }}Handler) Get{{ .EntityName }}(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'id' is required"})
		return
	}

	// Parse as UUID - only UUID is accepted
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	{{ .EntityNameLC }}, err := h.service.GetByUUID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "{{ .EntityName }} not found"})
		return
	}

	c.JSON(http.StatusOK, dto.{{ .EntityName }}Response{}.FromModel({{ .EntityNameLC }}, h.db))
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
		response[i] = dto.{{ .EntityName }}Response{}.FromModel({{ .EntityNameLC }}, h.db)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}

// Update{{ .EntityName }} updates a {{ toLower .EntityName }} by UUID (query param: id)
func (h *{{ .EntityName }}Handler) Update{{ .EntityName }}(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'id' is required"})
		return
	}

	// Parse as UUID - only UUID is accepted
	uuidVal, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var req dto.Update{{ .EntityName }}Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing record by UUID to get internal ID
	existing, err := h.service.GetByUUID(c.Request.Context(), uuidVal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "{{ .EntityName }} not found"})
		return
	}

	// Merge: only update fields that are provided (non-nil)
{{- range .Fields }}
{{- if .IsForeignKey }}
	// Translate FK UUID to internal ID
	if req.{{ .FKName }} != nil && *req.{{ .FKName }} != "" {
		{{ .FKNameLC }}, err := uuid.Parse(*req.{{ .FKName }})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid {{ .FKNameLC }} format"})
			return
		}
		var fkID int64
		err = h.db.Get(&fkID, "SELECT id FROM {{ toSnake (trimSuffix .Name "Id") }}s WHERE uuid = ? AND deleted_at IS NULL", {{ .FKNameLC }}.String())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "{{ .FKName }} not found"})
			return
		}
		existing.{{ .Name }} = fkID
	}
{{- else }}
	if req.{{ .Name }} != nil {
		existing.{{ .Name }} = *req.{{ .Name }}
	}
{{- end }}
{{- end }}

	// Update using internal ID
	if err := h.service.Update(c.Request.Context(), existing.ID, existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch updated record to return
	updated, err := h.service.GetByUUID(c.Request.Context(), uuidVal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.{{ .EntityName }}Response{}.FromModel(updated, h.db))
}

// Delete{{ .EntityName }} deletes a {{ toLower .EntityName }} by UUID (query param: id)
func (h *{{ .EntityName }}Handler) Delete{{ .EntityName }}(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'id' is required"})
		return
	}

	// Parse as UUID - only UUID is accepted
	uuidVal, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	// Get existing record by UUID to get internal ID
	existing, err := h.service.GetByUUID(c.Request.Context(), uuidVal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "{{ .EntityName }} not found"})
		return
	}

	// Delete using internal ID
	if err := h.service.Delete(c.Request.Context(), existing.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted"})
}
`

// DTO template
const dtoTemplate = `package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"{{ .ModuleName }}/models"
)

// Create{{ .EntityName }}Request represents a request to create a {{ toLower .EntityName }}
// NOTE: FK fields use UUID (e.g., category_uuid) - translation to internal ID happens in handler
type Create{{ .EntityName }}Request struct {
{{- range .Fields }}
{{- if .IsForeignKey }}
	{{ .FKName }} string ` + "`" + `{{ .FKJSONTag }}{{ if .ValidateTag }} {{ .ValidateTag }}{{ end }}` + "`" + ` // UUID of related entity
{{- else }}
	{{ .Name }} {{ .GoType }} ` + "`" + `{{ .JSONTag }}{{ if .ValidateTag }} {{ .ValidateTag }}{{ end }}` + "`" + `
{{- end }}
{{- end }}
}

// ToModel converts the request to a model (FK fields need UUID→ID translation in handler)
func (r Create{{ .EntityName }}Request) ToModel() *models.{{ .EntityName }} {
	return &models.{{ .EntityName }}{
{{- range .Fields }}
{{- if not .IsForeignKey }}
		{{ .Name }}: r.{{ .Name }},
{{- end }}
{{- end }}
	}
}

// Update{{ .EntityName }}Request represents a request to update a {{ toLower .EntityName }}
// NOTE: FK fields use UUID (e.g., category_uuid) - translation to internal ID happens in handler
// NOTE: All fields are optional pointers to support partial updates
type Update{{ .EntityName }}Request struct {
{{- range .Fields }}
{{- if .IsForeignKey }}
	{{ .FKName }} *string ` + "`" + `{{ .FKJSONTag }},omitempty{{ if .ValidateTagUpdate }} {{ .ValidateTagUpdate }}{{ end }}` + "`" + ` // UUID of related entity
{{- else }}
	{{ .Name }} *{{ .GoType }} ` + "`" + `json:"{{ toSnake .Name }},omitempty"{{ if .ValidateTagUpdate }} {{ .ValidateTagUpdate }}{{ end }}` + "`" + `
{{- end }}
{{- end }}
}

// ToModel converts the request to a model (FK fields need UUID→ID translation in handler)
// Only updates fields that are non-nil
func (r Update{{ .EntityName }}Request) ToModel() *models.{{ .EntityName }} {
	return &models.{{ .EntityName }}{
{{- range .Fields }}
{{- if not .IsForeignKey }}
{{- if eq .GoType "string" }}
		{{ .Name }}: func() string { if r.{{ .Name }} != nil { return *r.{{ .Name }} }; return "" }(),
{{- else if eq .GoType "int" }}
		{{ .Name }}: func() int { if r.{{ .Name }} != nil { return *r.{{ .Name }} }; return 0 }(),
{{- else if eq .GoType "int64" }}
		{{ .Name }}: func() int64 { if r.{{ .Name }} != nil { return *r.{{ .Name }} }; return 0 }(),
{{- else if eq .GoType "float64" }}
		{{ .Name }}: func() float64 { if r.{{ .Name }} != nil { return *r.{{ .Name }} }; return 0 }(),
{{- else if eq .GoType "bool" }}
		{{ .Name }}: func() bool { if r.{{ .Name }} != nil { return *r.{{ .Name }} }; return false }(),
{{- else }}
		{{ .Name }}: func() {{ .GoType }} { if r.{{ .Name }} != nil { return *r.{{ .Name }} }; return {{ .GoType }}{} }(),
{{- end }}
{{- end }}
{{- end }}
	}
}

// {{ .EntityName }}Response represents a {{ toLower .EntityName }} response
// NOTE: Only UUID is exposed - internal int64 ID is never exposed to frontend
// NOTE: FK fields are exposed as UUID (e.g., category_uuid instead of category_id)
type {{ .EntityName }}Response struct {
	UUID      uuid.UUID  ` + "`json:\"uuid\"`" + ` // UUID for frontend (internal int64 ID is hidden)
{{- range .Fields }}
{{- if .IsForeignKey }}
	{{ .FKName }} string ` + "`" + `{{ .FKJSONTag }}` + "`" + ` // UUID of related entity
{{- else }}
	{{ .Name }} {{ .GoType }} ` + "`" + `{{ .JSONTag }}` + "`" + `
{{- end }}
{{- end }}
	CreatedAt time.Time  ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time  ` + "`json:\"updated_at\"`" + `
	DeletedAt *time.Time ` + "`json:\"deleted_at,omitempty\"`" + `
}

// FromModel converts a model to a response
// NOTE: FK fields need ID→UUID translation (lookup related entity's UUID by ID)
func (r {{ .EntityName }}Response) FromModel({{ .EntityNameLC }} *models.{{ .EntityName }}, db *sqlx.DB) {{ .EntityName }}Response {
	var deletedAt *time.Time
	if {{ .EntityNameLC }}.DeletedAt.Valid {
		deletedAt = &{{ .EntityNameLC }}.DeletedAt.Time
	}

	response := {{ .EntityName }}Response{
		UUID:      {{ .EntityNameLC }}.UUID, // UUID for frontend, internal ID stays hidden
{{- range .Fields }}
{{- if not .IsForeignKey }}
		{{ .Name }}: {{ $.EntityNameLC }}.{{ .Name }},
{{- end }}
{{- end }}
		CreatedAt: {{ .EntityNameLC }}.CreatedAt,
		UpdatedAt: {{ .EntityNameLC }}.UpdatedAt,
		DeletedAt: deletedAt,
	}

	// Translate FK IDs to UUIDs
{{- range .Fields }}
{{- if .IsForeignKey }}
	if {{ $.EntityNameLC }}.{{ .Name }} != 0 {
		var fkUUID string
		err := db.Get(&fkUUID, "SELECT uuid FROM {{ toSnake (trimSuffix .Name "Id") }}s WHERE id = ?", {{ $.EntityNameLC }}.{{ .Name }})
		if err == nil {
			response.{{ .FKName }} = fkUUID
		}
	}
{{- end }}
{{- end }}

	return response
}
`

// Migration up template (MySQL)
const migrationUpTemplate = `-- Create {{ .TableName }} table
CREATE TABLE IF NOT EXISTS {{ .TableName }} (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) NOT NULL UNIQUE,
{{- range .Fields }}
    {{ toSnake .Name }} {{ sqlType .Type .Length }}{{ if .Required }} NOT NULL{{ end }}{{ if .Unique }} UNIQUE{{ end }}{{ if .DefaultValue }} DEFAULT {{ .DefaultValue }}{{ end }},
{{- end }}
{{- if .Relations }}
{{- range .Relations }}
{{- if or (eq .RelationType "belongsTo") (and (or (eq .RelationType "hasOne") (eq .RelationType "hasMany")) (not .IsSource)) }}
    {{ .FieldName }} BIGINT NOT NULL,
{{- end }}
{{- end }}
{{- end }}
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_{{ .TableName }}_uuid (uuid),
    INDEX idx_{{ .TableName }}_deleted_at (deleted_at),
    INDEX idx_{{ .TableName }}_created_at (created_at)
{{- if .Relations }}
{{- range .Relations }}
{{- if or (eq .RelationType "belongsTo") (and (or (eq .RelationType "hasOne") (eq .RelationType "hasMany")) (not .IsSource)) }}
    ,FOREIGN KEY ({{ .FieldName }}) REFERENCES {{ if .IsSource }}{{ .TargetTableName }}{{ else }}{{ .SourceTableName }}{{ end }}(id) ON DELETE {{ .OnDelete }} ON UPDATE {{ .OnUpdate }}
{{- end }}
{{- end }}
{{- end }}
);
`

// Migration down template
const migrationDownTemplate = `-- Drop {{ .TableName }} table
DROP TABLE IF EXISTS {{ .TableName }};
`
