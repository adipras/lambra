package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
	"github.com/yourusername/lambra/internal/utils"
)

type RelationService struct {
	relationRepo *repository.RelationRepository
	entityRepo   *repository.EntityRepository
	projectRepo  *repository.ProjectRepository
}

func NewRelationService(relationRepo *repository.RelationRepository, entityRepo *repository.EntityRepository, projectRepo *repository.ProjectRepository) *RelationService {
	return &RelationService{
		relationRepo: relationRepo,
		entityRepo:   entityRepo,
		projectRepo:  projectRepo,
	}
}

// CreateRelation creates a new relation
func (s *RelationService) CreateRelation(sourceEntityUUID, targetEntityUUID, fieldName, relationType, onDelete, onUpdate, junctionTable, description string) (*models.Relation, error) {
	// Get source entity
	sourceEntity, err := s.entityRepo.GetByUUID(sourceEntityUUID)
	if err != nil {
		return nil, fmt.Errorf("source entity not found: %w", err)
	}

	// Get target entity
	targetEntity, err := s.entityRepo.GetByUUID(targetEntityUUID)
	if err != nil {
		return nil, fmt.Errorf("target entity not found: %w", err)
	}

	// Auto-generate field name if not provided
	if fieldName == "" {
		fieldName = s.generateFieldName(targetEntity.Name, relationType)
	}

	// Validate field name doesn't conflict with existing fields
	if relationType != "manyToMany" {
		if err := s.validateFieldName(sourceEntity, targetEntity, fieldName, relationType); err != nil {
			return nil, err
		}
	}

	// Auto-generate junction table name for manyToMany
	if relationType == "manyToMany" && junctionTable == "" {
		junctionTable = s.generateJunctionTableName(sourceEntity.Name, targetEntity.Name)
	}

	// Create relation
	relation := &models.Relation{
		SourceEntityID:  sourceEntity.ID,
		SourceFieldName: fieldName,
		TargetEntityID:  targetEntity.ID,
		RelationType:    relationType,
		OnDelete:        onDelete,
		OnUpdate:        onUpdate,
	}

	if junctionTable != "" {
		relation.JunctionTable = sql.NullString{String: junctionTable, Valid: true}
	}
	if description != "" {
		relation.Description = sql.NullString{String: description, Valid: true}
	}

	// Validate relation
	if err := relation.Validate(); err != nil {
		return nil, err
	}

	// Check for duplicate relations
	existing, err := s.relationRepo.GetByEntities(sourceEntity.ID, targetEntity.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("relation already exists between these entities")
	}

	// Check for circular dependencies
	if err := s.checkCircularDependency(sourceEntity.ID, targetEntity.ID); err != nil {
		return nil, err
	}

	// Create in database
	if err := s.relationRepo.Create(relation); err != nil {
		return nil, err
	}

	// Fetch back with entity names
	return s.relationRepo.GetByUUID(relation.UUID)
}

// generateFieldName generates a foreign key field name
func (s *RelationService) generateFieldName(targetEntityName, relationType string) string {
	// Convert to snake_case and add _id
	fieldName := utils.ToSnakeCase(targetEntityName)

	// For belongsTo, the source has the FK
	if relationType == "belongsTo" {
		return fieldName + "_id"
	}

	// For hasOne/hasMany, target has FK to source (no field on source)
	// But we still need to store the field name for code generation
	return fieldName + "_id"
}

// generateJunctionTableName generates a junction table name for manyToMany
func (s *RelationService) generateJunctionTableName(source, target string) string {
	// Convert both to snake_case
	sourceName := utils.ToSnakeCase(utils.Pluralize(source))
	targetName := utils.ToSnakeCase(utils.Pluralize(target))

	// Sort alphabetically for consistency
	if sourceName < targetName {
		return sourceName + "_" + targetName
	}
	return targetName + "_" + sourceName
}

// checkCircularDependency checks if creating this relation would create a cycle
func (s *RelationService) checkCircularDependency(sourceID, targetID int64) error {
	// For now, simple check: prevent direct cycles (A->B and B->A)
	// TODO: Implement full graph traversal for complex cycle detection
	
	reverse, err := s.relationRepo.GetByEntities(targetID, sourceID)
	if err != nil {
		return err
	}
	if reverse != nil {
		return errors.New("circular dependency detected: reverse relation already exists")
	}
	
	return nil
}

// GetRelationByUUID retrieves a relation by UUID
func (s *RelationService) GetRelationByUUID(uuid string) (*models.Relation, error) {
	return s.relationRepo.GetByUUID(uuid)
}

// GetEntityRelations retrieves all relations for an entity (both source and target)
func (s *RelationService) GetEntityRelations(entityUUID string) ([]models.Relation, error) {
	entity, err := s.entityRepo.GetByUUID(entityUUID)
	if err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}

	// Get relations where entity is source
	sourceRelations, err := s.relationRepo.GetBySourceEntity(entity.ID)
	if err != nil {
		return nil, err
	}

	// Get relations where entity is target
	targetRelations, err := s.relationRepo.GetByTargetEntity(entity.ID)
	if err != nil {
		return nil, err
	}

	// Combine both
	allRelations := append(sourceRelations, targetRelations...)
	return allRelations, nil
}

// GetProjectRelations retrieves all relations for a project
func (s *RelationService) GetProjectRelations(projectUUID string) ([]models.Relation, error) {
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}
	return s.relationRepo.GetByProject(project.ID)
}

// UpdateRelation updates a relation
func (s *RelationService) UpdateRelation(uuid, fieldName, relationType, onDelete, onUpdate, junctionTable, description string) (*models.Relation, error) {
	// Get existing relation
	existing, err := s.relationRepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}

	// Update fields
	if fieldName != "" {
		existing.SourceFieldName = fieldName
	}
	if relationType != "" {
		existing.RelationType = relationType
	}
	if onDelete != "" {
		existing.OnDelete = onDelete
	}
	if onUpdate != "" {
		existing.OnUpdate = onUpdate
	}
	if junctionTable != "" {
		existing.JunctionTable = sql.NullString{String: junctionTable, Valid: true}
	}
	if description != "" {
		existing.Description = sql.NullString{String: description, Valid: true}
	}

	// Validate
	if err := existing.Validate(); err != nil {
		return nil, err
	}

	// Update in database
	if err := s.relationRepo.Update(uuid, existing); err != nil {
		return nil, err
	}

	// Fetch back with entity names
	return s.relationRepo.GetByUUID(uuid)
}

// DeleteRelation soft deletes a relation
func (s *RelationService) DeleteRelation(uuid, deletedBy string) error {
	return s.relationRepo.DeleteByUUID(uuid, deletedBy)
}

// DeleteEntityRelations deletes all relations for an entity (when entity is deleted)
func (s *RelationService) DeleteEntityRelations(entityUUID, deletedBy string) error {
	entity, err := s.entityRepo.GetByUUID(entityUUID)
	if err != nil {
		return err
	}

	// Delete relations where entity is source
	if err := s.relationRepo.DeleteBySourceEntity(entity.ID, deletedBy); err != nil {
		return err
	}

	// Delete relations where entity is target
	if err := s.relationRepo.DeleteByTargetEntity(entity.ID, deletedBy); err != nil {
		return err
	}

	return nil
}

// validateFieldName validates that FK field name doesn't conflict with existing fields
func (s *RelationService) validateFieldName(sourceEntity, targetEntity *models.Entity, fieldName, relationType string) error {
	// Determine which entity will have the FK column
	var entityWithFK *models.Entity
	var entityName string

	switch relationType {
	case "belongsTo":
		// FK in source entity
		entityWithFK = sourceEntity
		entityName = "source"
	case "hasOne", "hasMany":
		// FK in target entity
		entityWithFK = targetEntity
		entityName = "target"
	default:
		return nil // No validation needed for manyToMany
	}

	// Parse entity fields
	var fields []models.EntityField
	if err := json.Unmarshal(entityWithFK.Fields, &fields); err != nil {
		return fmt.Errorf("failed to parse entity fields: %w", err)
	}

	// Convert field name to snake_case for comparison
	fieldNameSnake := utils.ToSnakeCase(fieldName)

	// Check if field name conflicts with existing fields
	for _, field := range fields {
		existingFieldName := utils.ToSnakeCase(field.Name)
		if existingFieldName == fieldNameSnake {
			// Field exists - check type compatibility
			// int/int64 types can be converted to BIGINT (OK)
			// Other types will be overridden (WARN but allow)
			if field.Type != "int" && field.Type != "int64" {
				// Log warning but allow
				fmt.Printf("WARNING: Field '%s' in %s entity (%s) has type %s but will be converted to BIGINT for FK\n",
					fieldName, entityName, entityWithFK.Name, field.Type)
			}
			// Allow - field will be overridden in migration
			return nil
		}
	}

	// Field doesn't exist - will be created as new field
	// Validate field name format
	if len(fieldName) < 2 {
		return errors.New("field name too short (minimum 2 characters)")
	}

	// All good - new field will be created
	return nil
}
