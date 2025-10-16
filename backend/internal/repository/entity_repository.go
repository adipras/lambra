package repository

import (
	"database/sql"
	"fmt"
	"math/big"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/internal/models"
)

type EntityRepository struct {
	db *sqlx.DB
}

func NewEntityRepository(db *sqlx.DB) *EntityRepository {
	return &EntityRepository{db: db}
}

// uuidToInt64 converts UUID to int64 by taking first 8 bytes
func uuidToInt64Entity(u uuid.UUID) int64 {
	bytes := u[:]
	var num big.Int
	num.SetBytes(bytes[:8])
	return num.Int64()
}

func (r *EntityRepository) Create(entity *models.Entity) error {
	// Generate UUID v7
	uuidV7 := uuid.Must(uuid.NewV7())
	id := uuidToInt64Entity(uuidV7)
	uuidStr := uuidV7.String()

	query := `
		INSERT INTO entities (id, uuid, project_id, name, table_name, description, fields, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	_, err := r.db.Exec(query, id, uuidStr, entity.ProjectID, entity.Name, entity.TableName, entity.Description, entity.Fields, entity.CreatedBy)
	if err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}

	// Get the created entity to populate all fields including timestamps
	createdEntity, err := r.GetByUUID(uuidStr)
	if err != nil {
		return fmt.Errorf("failed to retrieve created entity: %w", err)
	}

	*entity = *createdEntity
	return nil
}

// GetByUUID retrieves entity by UUID (external identifier)
func (r *EntityRepository) GetByUUID(uuid string) (*models.Entity, error) {
	var entity models.Entity
	query := `
		SELECT id, uuid, project_id, name, table_name, description, fields,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM entities
		WHERE uuid = ? AND deleted_at IS NULL
	`

	err := r.db.Get(&entity, query, uuid)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("entity not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	return &entity, nil
}

// GetByID retrieves entity by internal ID (for internal use/FK joins)
func (r *EntityRepository) GetByID(id int64) (*models.Entity, error) {
	var entity models.Entity
	query := `
		SELECT id, uuid, project_id, name, table_name, description, fields,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM entities
		WHERE id = ? AND deleted_at IS NULL
	`

	err := r.db.Get(&entity, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("entity not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	return &entity, nil
}

func (r *EntityRepository) GetByProjectID(projectID int64) ([]models.Entity, error) {
	var entities []models.Entity
	query := `
		SELECT id, uuid, project_id, name, table_name, description, fields,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM entities
		WHERE project_id = ? AND deleted_at IS NULL
		ORDER BY created_at ASC
	`

	err := r.db.Select(&entities, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	return entities, nil
}

func (r *EntityRepository) Update(entity *models.Entity) error {
	query := `
		UPDATE entities
		SET name = ?, table_name = ?, description = ?, fields = ?, updated_by = ?, updated_at = NOW()
		WHERE uuid = ? AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query, entity.Name, entity.TableName, entity.Description, entity.Fields, entity.UpdatedBy, entity.UUID)
	if err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}

	return nil
}

func (r *EntityRepository) DeleteByUUID(uuid string, deletedBy string) error {
	// Soft delete
	query := `UPDATE entities SET deleted_by = ?, deleted_at = NOW() WHERE uuid = ? AND deleted_at IS NULL`
	_, err := r.db.Exec(query, deletedBy, uuid)
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	return nil
}
