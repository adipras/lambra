package repository

import (
	"database/sql"
	"fmt"
	"math/big"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/internal/models"
)

type EndpointRepository struct {
	db *sqlx.DB
}

func NewEndpointRepository(db *sqlx.DB) *EndpointRepository {
	return &EndpointRepository{db: db}
}

// uuidToInt64 converts UUID to int64 by taking first 8 bytes
func uuidToInt64Endpoint(u uuid.UUID) int64 {
	bytes := u[:]
	var num big.Int
	num.SetBytes(bytes[:8])
	return num.Int64()
}

func (r *EndpointRepository) Create(endpoint *models.Endpoint) error {
	// Generate UUID v7
	uuidV7 := uuid.Must(uuid.NewV7())
	id := uuidToInt64Endpoint(uuidV7)
	uuidStr := uuidV7.String()

	query := `
		INSERT INTO endpoints (id, uuid, entity_id, project_id, name, path, method, description,
								request_schema, response_schema, require_auth, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	_, err := r.db.Exec(query, id, uuidStr,
		endpoint.EntityID, endpoint.ProjectID, endpoint.Name, endpoint.Path, endpoint.Method,
		endpoint.Description, endpoint.RequestSchema, endpoint.ResponseSchema, endpoint.RequireAuth, endpoint.CreatedBy)
	if err != nil {
		return fmt.Errorf("failed to create endpoint: %w", err)
	}

	// Get the created endpoint to populate all fields including timestamps
	createdEndpoint, err := r.GetByUUID(uuidStr)
	if err != nil {
		return fmt.Errorf("failed to retrieve created endpoint: %w", err)
	}

	*endpoint = *createdEndpoint
	return nil
}

// GetByUUID retrieves endpoint by UUID (external identifier)
func (r *EndpointRepository) GetByUUID(uuid string) (*models.Endpoint, error) {
	var endpoint models.Endpoint
	query := `
		SELECT id, uuid, entity_id, project_id, name, path, method, description,
		       request_schema, response_schema, require_auth,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM endpoints
		WHERE uuid = ? AND deleted_at IS NULL
	`

	err := r.db.Get(&endpoint, query, uuid)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("endpoint not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint: %w", err)
	}

	return &endpoint, nil
}

// GetByID retrieves endpoint by internal ID (for internal use/FK joins)
func (r *EndpointRepository) GetByID(id int64) (*models.Endpoint, error) {
	var endpoint models.Endpoint
	query := `
		SELECT id, uuid, entity_id, project_id, name, path, method, description,
		       request_schema, response_schema, require_auth,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM endpoints
		WHERE id = ? AND deleted_at IS NULL
	`

	err := r.db.Get(&endpoint, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("endpoint not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint: %w", err)
	}

	return &endpoint, nil
}

func (r *EndpointRepository) GetByProjectID(projectID int64) ([]models.Endpoint, error) {
	var endpoints []models.Endpoint
	query := `
		SELECT id, uuid, entity_id, project_id, name, path, method, description,
		       request_schema, response_schema, require_auth,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM endpoints
		WHERE project_id = ? AND deleted_at IS NULL
		ORDER BY created_at ASC
	`

	err := r.db.Select(&endpoints, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints: %w", err)
	}

	return endpoints, nil
}

func (r *EndpointRepository) GetByEntityID(entityID int64) ([]models.Endpoint, error) {
	var endpoints []models.Endpoint
	query := `
		SELECT id, uuid, entity_id, project_id, name, path, method, description,
		       request_schema, response_schema, require_auth,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM endpoints
		WHERE entity_id = ? AND deleted_at IS NULL
		ORDER BY created_at ASC
	`

	err := r.db.Select(&endpoints, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints: %w", err)
	}

	return endpoints, nil
}

func (r *EndpointRepository) Update(endpoint *models.Endpoint) error {
	query := `
		UPDATE endpoints
		SET name = ?, path = ?, method = ?, description = ?,
			request_schema = ?, response_schema = ?, require_auth = ?, updated_by = ?, updated_at = NOW()
		WHERE uuid = ? AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query,
		endpoint.Name, endpoint.Path, endpoint.Method, endpoint.Description,
		endpoint.RequestSchema, endpoint.ResponseSchema, endpoint.RequireAuth, endpoint.UpdatedBy, endpoint.UUID)
	if err != nil {
		return fmt.Errorf("failed to update endpoint: %w", err)
	}

	return nil
}

func (r *EndpointRepository) DeleteByUUID(uuid string, deletedBy string) error {
	// Soft delete
	query := `UPDATE endpoints SET deleted_by = ?, deleted_at = NOW() WHERE uuid = ? AND deleted_at IS NULL`
	_, err := r.db.Exec(query, deletedBy, uuid)
	if err != nil {
		return fmt.Errorf("failed to delete endpoint: %w", err)
	}

	return nil
}
