package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/yourusername/lambra/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RelationRepository struct {
	db *sqlx.DB
}

func NewRelationRepository(db *sqlx.DB) *RelationRepository {
	return &RelationRepository{db: db}
}

// Create creates a new relation
func (r *RelationRepository) Create(relation *models.Relation) error {
	// Generate UUID v7 and derive int64 ID
	uuidV7 := uuid.Must(uuid.NewV7())
	relation.ID = uuidToInt64(uuidV7)
	relation.UUID = uuidV7.String()
	relation.CreatedAt = time.Now()
	relation.UpdatedAt = time.Now()

	query := `
		INSERT INTO relations (
			id, uuid, source_entity_id, source_field_name, target_entity_id,
			relation_type, on_delete, on_update, junction_table, description,
			created_by, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		relation.ID,
		relation.UUID,
		relation.SourceEntityID,
		relation.SourceFieldName,
		relation.TargetEntityID,
		relation.RelationType,
		relation.OnDelete,
		relation.OnUpdate,
		relation.JunctionTable,
		relation.Description,
		relation.CreatedBy,
		relation.CreatedAt,
		relation.UpdatedAt,
	)

	return err
}

// GetByID retrieves a relation by internal ID
func (r *RelationRepository) GetByID(id int64) (*models.Relation, error) {
	var relation models.Relation
	query := `
		SELECT r.*,
		       se.uuid as source_entity_uuid, se.name as source_entity_name,
		       te.uuid as target_entity_uuid, te.name as target_entity_name
		FROM relations r
		INNER JOIN entities se ON r.source_entity_id = se.id
		INNER JOIN entities te ON r.target_entity_id = te.id
		WHERE r.id = ? AND r.deleted_at IS NULL
	`
	err := r.db.Get(&relation, query, id)
	if err != nil {
		return nil, err
	}
	return &relation, nil
}

// GetByUUID retrieves a relation by UUID
func (r *RelationRepository) GetByUUID(uuid string) (*models.Relation, error) {
	var relation models.Relation
	query := `
		SELECT r.*,
		       se.uuid as source_entity_uuid, se.name as source_entity_name,
		       te.uuid as target_entity_uuid, te.name as target_entity_name
		FROM relations r
		INNER JOIN entities se ON r.source_entity_id = se.id
		INNER JOIN entities te ON r.target_entity_id = te.id
		WHERE r.uuid = ? AND r.deleted_at IS NULL
	`
	err := r.db.Get(&relation, query, uuid)
	if err != nil {
		return nil, err
	}
	return &relation, nil
}

// GetBySourceEntity retrieves all relations where entity is the source
func (r *RelationRepository) GetBySourceEntity(entityID int64) ([]models.Relation, error) {
	var relations []models.Relation
	query := `
		SELECT r.*,
		       se.uuid as source_entity_uuid, se.name as source_entity_name,
		       te.uuid as target_entity_uuid, te.name as target_entity_name
		FROM relations r
		INNER JOIN entities se ON r.source_entity_id = se.id
		INNER JOIN entities te ON r.target_entity_id = te.id
		WHERE r.source_entity_id = ? AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC
	`
	err := r.db.Select(&relations, query, entityID)
	if err != nil {
		return nil, err
	}
	return relations, nil
}

// GetByTargetEntity retrieves all relations where entity is the target
func (r *RelationRepository) GetByTargetEntity(entityID int64) ([]models.Relation, error) {
	var relations []models.Relation
	query := `
		SELECT r.*,
		       se.uuid as source_entity_uuid, se.name as source_entity_name,
		       te.uuid as target_entity_uuid, te.name as target_entity_name
		FROM relations r
		INNER JOIN entities se ON r.source_entity_id = se.id
		INNER JOIN entities te ON r.target_entity_id = te.id
		WHERE r.target_entity_id = ? AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC
	`
	err := r.db.Select(&relations, query, entityID)
	if err != nil {
		return nil, err
	}
	return relations, nil
}

// GetByEntities checks if a relation exists between two entities
func (r *RelationRepository) GetByEntities(sourceID, targetID int64) (*models.Relation, error) {
	var relation models.Relation
	query := `
		SELECT r.*,
		       se.uuid as source_entity_uuid, se.name as source_entity_name,
		       te.uuid as target_entity_uuid, te.name as target_entity_name
		FROM relations r
		INNER JOIN entities se ON r.source_entity_id = se.id
		INNER JOIN entities te ON r.target_entity_id = te.id
		WHERE r.source_entity_id = ? AND r.target_entity_id = ?
		  AND r.deleted_at IS NULL
		LIMIT 1
	`
	err := r.db.Get(&relation, query, sourceID, targetID)
	if err == sql.ErrNoRows {
		return nil, nil // No relation found, not an error
	}
	if err != nil {
		return nil, err
	}
	return &relation, nil
}

// GetByProject retrieves all relations for entities in a project
func (r *RelationRepository) GetByProject(projectID int64) ([]models.Relation, error) {
	var relations []models.Relation
	query := `
		SELECT r.*,
		       se.uuid as source_entity_uuid, se.name as source_entity_name,
		       te.uuid as target_entity_uuid, te.name as target_entity_name
		FROM relations r
		INNER JOIN entities se ON r.source_entity_id = se.id
		INNER JOIN entities te ON r.target_entity_id = te.id
		WHERE se.project_id = ? AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC
	`
	err := r.db.Select(&relations, query, projectID)
	if err != nil {
		return nil, err
	}
	return relations, nil
}

// List retrieves all relations
func (r *RelationRepository) List(limit, offset int) ([]models.Relation, error) {
	var relations []models.Relation
	query := `
		SELECT r.*,
		       se.uuid as source_entity_uuid, se.name as source_entity_name,
		       te.uuid as target_entity_uuid, te.name as target_entity_name
		FROM relations r
		INNER JOIN entities se ON r.source_entity_id = se.id
		INNER JOIN entities te ON r.target_entity_id = te.id
		WHERE r.deleted_at IS NULL
		ORDER BY r.created_at DESC
		LIMIT ? OFFSET ?
	`
	err := r.db.Select(&relations, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return relations, nil
}

// Update updates a relation
func (r *RelationRepository) Update(uuid string, relation *models.Relation) error {
	relation.UpdatedAt = time.Now()

	query := `
		UPDATE relations
		SET source_field_name = ?,
		    relation_type = ?,
		    on_delete = ?,
		    on_update = ?,
		    junction_table = ?,
		    description = ?,
		    updated_by = ?,
		    updated_at = ?
		WHERE uuid = ? AND deleted_at IS NULL
	`

	result, err := r.db.Exec(
		query,
		relation.SourceFieldName,
		relation.RelationType,
		relation.OnDelete,
		relation.OnUpdate,
		relation.JunctionTable,
		relation.Description,
		relation.UpdatedBy,
		relation.UpdatedAt,
		uuid,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("relation not found or already deleted")
	}

	return nil
}

// DeleteByUUID soft deletes a relation
func (r *RelationRepository) DeleteByUUID(uuid string, deletedBy string) error {
	query := `
		UPDATE relations
		SET deleted_at = ?,
		    deleted_by = ?
		WHERE uuid = ? AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, time.Now(), deletedBy, uuid)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("relation not found or already deleted")
	}

	return nil
}

// DeleteBySourceEntity soft deletes all relations for a source entity
func (r *RelationRepository) DeleteBySourceEntity(entityID int64, deletedBy string) error {
	query := `
		UPDATE relations
		SET deleted_at = ?,
		    deleted_by = ?
		WHERE source_entity_id = ? AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query, time.Now(), deletedBy, entityID)
	return err
}

// DeleteByTargetEntity soft deletes all relations for a target entity
func (r *RelationRepository) DeleteByTargetEntity(entityID int64, deletedBy string) error {
	query := `
		UPDATE relations
		SET deleted_at = ?,
		    deleted_by = ?
		WHERE target_entity_id = ? AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query, time.Now(), deletedBy, entityID)
	return err
}

// SoftDeleteByEntity soft deletes all relations for an entity (both source and target)
func (r *RelationRepository) SoftDeleteByEntity(entityID int64, deletedBy string) error {
	query := `
		UPDATE relations
		SET deleted_at = ?,
		    deleted_by = ?
		WHERE (source_entity_id = ? OR target_entity_id = ?) AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query, time.Now(), deletedBy, entityID, entityID)
	return err
}

// Count returns the total number of relations
func (r *RelationRepository) Count() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM relations WHERE deleted_at IS NULL`
	err := r.db.Get(&count, query)
	return count, err
}
