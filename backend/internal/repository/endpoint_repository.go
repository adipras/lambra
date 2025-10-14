package repository

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/internal/models"
)

type EndpointRepository struct {
	db *sqlx.DB
}

func NewEndpointRepository(db *sqlx.DB) *EndpointRepository {
	return &EndpointRepository{db: db}
}

func (r *EndpointRepository) Create(endpoint *models.Endpoint) error {
	query := `
		INSERT INTO endpoints (entity_id, project_id, name, path, method, description,
								request_schema, response_schema, require_auth, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := r.db.Exec(query,
		endpoint.EntityID, endpoint.ProjectID, endpoint.Name, endpoint.Path, endpoint.Method,
		endpoint.Description, endpoint.RequestSchema, endpoint.ResponseSchema, endpoint.RequireAuth)
	if err != nil {
		return fmt.Errorf("failed to create endpoint: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	endpoint.ID = id
	return nil
}

func (r *EndpointRepository) GetByID(id int64) (*models.Endpoint, error) {
	var endpoint models.Endpoint
	query := `SELECT * FROM endpoints WHERE id = ?`

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
	query := `SELECT * FROM endpoints WHERE project_id = ? ORDER BY created_at ASC`

	err := r.db.Select(&endpoints, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints: %w", err)
	}

	return endpoints, nil
}

func (r *EndpointRepository) GetByEntityID(entityID int64) ([]models.Endpoint, error) {
	var endpoints []models.Endpoint
	query := `SELECT * FROM endpoints WHERE entity_id = ? ORDER BY created_at ASC`

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
			request_schema = ?, response_schema = ?, require_auth = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err := r.db.Exec(query,
		endpoint.Name, endpoint.Path, endpoint.Method, endpoint.Description,
		endpoint.RequestSchema, endpoint.ResponseSchema, endpoint.RequireAuth, endpoint.ID)
	if err != nil {
		return fmt.Errorf("failed to update endpoint: %w", err)
	}

	return nil
}

func (r *EndpointRepository) Delete(id int64) error {
	query := `DELETE FROM endpoints WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete endpoint: %w", err)
	}

	return nil
}
