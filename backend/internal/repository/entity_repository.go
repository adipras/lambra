package repository

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/internal/models"
)

type EntityRepository struct {
	db *sqlx.DB
}

func NewEntityRepository(db *sqlx.DB) *EntityRepository {
	return &EntityRepository{db: db}
}

func (r *EntityRepository) Create(entity *models.Entity) error {
	query := `
		INSERT INTO entities (project_id, name, table_name, description, fields, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := r.db.Exec(query, entity.ProjectID, entity.Name, entity.TableName, entity.Description, entity.Fields)
	if err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	entity.ID = id
	return nil
}

func (r *EntityRepository) GetByID(id int64) (*models.Entity, error) {
	var entity models.Entity
	query := `SELECT * FROM entities WHERE id = ?`

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
	query := `SELECT * FROM entities WHERE project_id = ? ORDER BY created_at ASC`

	err := r.db.Select(&entities, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	return entities, nil
}

func (r *EntityRepository) Update(entity *models.Entity) error {
	query := `
		UPDATE entities
		SET name = ?, table_name = ?, description = ?, fields = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err := r.db.Exec(query, entity.Name, entity.TableName, entity.Description, entity.Fields, entity.ID)
	if err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}

	return nil
}

func (r *EntityRepository) Delete(id int64) error {
	query := `DELETE FROM entities WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	return nil
}
