package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
)

type SnapshotService struct {
	snapshotRepo *repository.SnapshotRepository
	projectRepo  *repository.ProjectRepository
	entityRepo   *repository.EntityRepository
	endpointRepo *repository.EndpointRepository
	relationRepo *repository.RelationRepository
}

func NewSnapshotService(
	snapshotRepo *repository.SnapshotRepository,
	projectRepo *repository.ProjectRepository,
	entityRepo *repository.EntityRepository,
	endpointRepo *repository.EndpointRepository,
	relationRepo *repository.RelationRepository,
) *SnapshotService {
	return &SnapshotService{
		snapshotRepo: snapshotRepo,
		projectRepo:  projectRepo,
		entityRepo:   entityRepo,
		endpointRepo: endpointRepo,
		relationRepo: relationRepo,
	}
}

// CreateSnapshot creates a new snapshot for a project
func (s *SnapshotService) CreateSnapshot(projectUUID string, createdBy string) (*models.GenerationSnapshot, error) {
	// Get project
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Get all entities for the project
	entities, err := s.entityRepo.GetByProjectID(project.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	// Get all relations for the project
	relations, err := s.relationRepo.GetByProject(project.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get relations: %w", err)
	}

	// Get all endpoints for each entity with entity name for mapping during rollback
	var allEndpoints []models.SnapshotEndpoint
	for _, entity := range entities {
		endpoints, err := s.endpointRepo.GetByEntityID(entity.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get endpoints for entity %s: %w", entity.Name, err)
		}
		// Wrap each endpoint with entity name for proper mapping during rollback
		for _, ep := range endpoints {
			allEndpoints = append(allEndpoints, models.SnapshotEndpoint{
				Endpoint:   ep,
				EntityName: entity.Name,
			})
		}
	}

	// Create snapshot metadata
	metadata := models.SnapshotMetadata{
		Entities:  entities,
		Endpoints: allEndpoints,
		Relations: relations,
		Config: map[string]interface{}{
			"namespace": project.Namespace,
			"db_host":   project.DBHost,
			"db_port":   project.DBPort,
			"db_user":   project.DBUser,
			"db_name":   project.DBName,
		},
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create database snapshot info
	dbSnapshot := models.DatabaseSnapshotInfo{
		MigrationVersion:  fmt.Sprintf("v%s", time.Now().Format("20060102150405")),
		AppliedMigrations: []string{},
	}

	dbSnapshotJSON, err := json.Marshal(dbSnapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal database snapshot: %w", err)
	}

	// Count existing snapshots for version numbering
	count, err := s.snapshotRepo.CountByProjectID(project.ID)
	if err != nil {
		count = 0
	}

	// Generate version
	version := fmt.Sprintf("v1.%d.0", count+1)

	// Deactivate previous active snapshot
	if err := s.snapshotRepo.SetAllInactiveByProjectID(project.ID, createdBy); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to deactivate old snapshots: %v\n", err)
	}

	// Create snapshot
	snapshot := &models.GenerationSnapshot{
		ProjectID:        project.ID,
		Version:          version,
		GitCommitHash:    fmt.Sprintf("local-%d", time.Now().Unix()),
		Metadata:         metadataJSON,
		DatabaseSnapshot: dbSnapshotJSON,
		Status:           models.SnapshotStatusActive,
	}
	snapshot.CreatedBy.String = createdBy
	snapshot.CreatedBy.Valid = true

	if err := s.snapshotRepo.Create(snapshot); err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Populate ProjectUUID for response
	snapshot.ProjectUUID = project.UUID

	return snapshot, nil
}

// GetSnapshot retrieves a snapshot by UUID
func (s *SnapshotService) GetSnapshot(snapshotUUID string) (*models.GenerationSnapshot, error) {
	snapshot, err := s.snapshotRepo.GetByUUID(snapshotUUID)
	if err != nil {
		return nil, fmt.Errorf("snapshot not found: %w", err)
	}

	// Populate ProjectUUID
	project, err := s.projectRepo.GetByID(snapshot.ProjectID)
	if err == nil {
		snapshot.ProjectUUID = project.UUID
	}

	return snapshot, nil
}

// GetSnapshotsByProject retrieves all snapshots for a project
func (s *SnapshotService) GetSnapshotsByProject(projectUUID string, limit, offset int) ([]models.GenerationSnapshot, int64, error) {
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, 0, fmt.Errorf("project not found: %w", err)
	}

	snapshots, total, err := s.snapshotRepo.GetByProjectID(project.ID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get snapshots: %w", err)
	}

	// Populate ProjectUUID for all snapshots
	for i := range snapshots {
		snapshots[i].ProjectUUID = project.UUID
	}

	return snapshots, total, nil
}

// GetLatestSnapshot retrieves the latest snapshot for a project
func (s *SnapshotService) GetLatestSnapshot(projectUUID string) (*models.GenerationSnapshot, error) {
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	snapshot, err := s.snapshotRepo.GetLatestByProjectID(project.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest snapshot: %w", err)
	}

	if snapshot == nil {
		return nil, nil
	}

	snapshot.ProjectUUID = project.UUID
	return snapshot, nil
}

// inferEntityNameFromPath tries to match an endpoint path to an entity name
// Used for backward compatibility with old snapshots that don't have entity_name field
// Example: /categories → Category, /units → Unit
func inferEntityNameFromPath(path string, entities []models.Entity) string {
	// Clean the path and get the first segment
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}

	pathBase := strings.ToLower(parts[0])

	// Try to match against entity table names (which are usually plural)
	for _, entity := range entities {
		tableName := strings.ToLower(entity.TableName)
		if tableName == pathBase || tableName+"s" == pathBase {
			return entity.Name
		}
	}

	// Try to match against entity names (convert to snake_case plural)
	for _, entity := range entities {
		entityLower := strings.ToLower(entity.Name)
		// Simple pluralization check
		if entityLower+"s" == pathBase || entityLower+"es" == pathBase ||
			entityLower == pathBase {
			return entity.Name
		}
		// Handle -y → -ies (category → categories)
		if strings.HasSuffix(entityLower, "y") {
			plural := entityLower[:len(entityLower)-1] + "ies"
			if plural == pathBase {
				return entity.Name
			}
		}
	}

	return ""
}

// RollbackToSnapshot restores entities and endpoints from a snapshot
// Returns the project UUID for redeployment
func (s *SnapshotService) RollbackToSnapshot(snapshotUUID string, rolledBackBy string) (string, error) {
	// Get snapshot
	snapshot, err := s.snapshotRepo.GetByUUID(snapshotUUID)
	if err != nil {
		return "", fmt.Errorf("snapshot not found: %w", err)
	}

	// Get project
	project, err := s.projectRepo.GetByID(snapshot.ProjectID)
	if err != nil {
		return "", fmt.Errorf("project not found: %w", err)
	}

	// Parse metadata
	var metadata models.SnapshotMetadata
	if err := json.Unmarshal(snapshot.Metadata, &metadata); err != nil {
		return "", fmt.Errorf("failed to parse snapshot metadata: %w", err)
	}

	// Soft delete all current entities, endpoints, and relations
	// Note: Application-level soft deletes don't cascade via FK, so we must explicitly delete
	currentEntities, err := s.entityRepo.GetByProjectID(project.ID)
	if err == nil {
		for _, entity := range currentEntities {
			// First soft-delete endpoints for this entity
			s.endpointRepo.SoftDeleteByEntityID(entity.ID, rolledBackBy)
			// Then soft-delete relations for this entity
			s.relationRepo.SoftDeleteByEntity(entity.ID, rolledBackBy)
			// Then soft-delete the entity itself
			s.entityRepo.DeleteByUUID(entity.UUID, rolledBackBy)
		}
	}

	// Restore entities from snapshot
	// Use entity name as key (unique per project) since entity.ID is not serialized to JSON
	entityNameToIDMap := make(map[string]int64) // entity name -> new ID
	for _, entity := range metadata.Entities {
		newEntity := &models.Entity{
			ProjectID:   project.ID,
			Name:        entity.Name,
			TableName:   entity.TableName,
			Description: entity.Description,
			Fields:      entity.Fields,
		}
		newEntity.SetCreatedBy(rolledBackBy)

		if err := s.entityRepo.Create(newEntity); err != nil {
			return "", fmt.Errorf("failed to restore entity %s: %w", entity.Name, err)
		}

		entityNameToIDMap[entity.Name] = newEntity.ID
	}

	// Restore endpoints from snapshot
	// Use EntityName from SnapshotEndpoint for proper mapping
	// For backward compatibility with old snapshots that don't have entity_name,
	// try to infer entity from endpoint path pattern (e.g., /categories → Category)
	for _, snapshotEndpoint := range metadata.Endpoints {
		entityName := snapshotEndpoint.EntityName

		// Backward compatibility: if EntityName is empty, try to infer from path
		if entityName == "" {
			entityName = inferEntityNameFromPath(snapshotEndpoint.Path, metadata.Entities)
			if entityName == "" {
				fmt.Printf("Warning: could not determine entity for endpoint %s (path: %s), skipping\n",
					snapshotEndpoint.Name, snapshotEndpoint.Path)
				continue
			}
			fmt.Printf("Info: inferred entity '%s' for endpoint %s from path\n", entityName, snapshotEndpoint.Name)
		}

		newEntityID, ok := entityNameToIDMap[entityName]
		if !ok {
			fmt.Printf("Warning: entity %s not found for endpoint %s, skipping\n", entityName, snapshotEndpoint.Name)
			continue // Skip if entity wasn't restored
		}

		newEndpoint := &models.Endpoint{
			EntityID:        newEntityID,
			ProjectID:       project.ID,
			Name:            snapshotEndpoint.Name,
			Path:            snapshotEndpoint.Path,
			Method:          snapshotEndpoint.Method,
			Description:     snapshotEndpoint.Description,
			RequestSchema:   snapshotEndpoint.RequestSchema,
			ResponseSchema:  snapshotEndpoint.ResponseSchema,
			RequireAuth:     snapshotEndpoint.RequireAuth,
		}
		newEndpoint.SetCreatedBy(rolledBackBy)

		if err := s.endpointRepo.Create(newEndpoint); err != nil {
			fmt.Printf("Warning: failed to restore endpoint %s: %v\n", snapshotEndpoint.Name, err)
		}
	}

	// Restore relations from snapshot
	// Need to map old entity IDs to new entity IDs
	for _, relation := range metadata.Relations {
		// Find source entity by name
		var sourceEntityID int64
		for name, id := range entityNameToIDMap {
			// Match by entity name (find the entity with matching ID from metadata)
			for _, entity := range metadata.Entities {
				if entity.ID == relation.SourceEntityID && entity.Name == name {
					sourceEntityID = id
					break
				}
			}
			if sourceEntityID != 0 {
				break
			}
		}

		// Find target entity by name
		var targetEntityID int64
		for name, id := range entityNameToIDMap {
			for _, entity := range metadata.Entities {
				if entity.ID == relation.TargetEntityID && entity.Name == name {
					targetEntityID = id
					break
				}
			}
			if targetEntityID != 0 {
				break
			}
		}

		if sourceEntityID == 0 || targetEntityID == 0 {
			fmt.Printf("Warning: could not map relation %s, skipping\n", relation.UUID)
			continue
		}

		newRelation := &models.Relation{
			SourceEntityID:  sourceEntityID,
			TargetEntityID:  targetEntityID,
			RelationType:    relation.RelationType,
			SourceFieldName: relation.SourceFieldName,
			OnDelete:        relation.OnDelete,
			OnUpdate:        relation.OnUpdate,
			JunctionTable:   relation.JunctionTable,
			Description:     relation.Description,
			Required:        relation.Required,
		}
		newRelation.SetCreatedBy(rolledBackBy)

		if err := s.relationRepo.Create(newRelation); err != nil {
			fmt.Printf("Warning: failed to restore relation: %v\n", err)
		}
	}

	// Update snapshot statuses
	s.snapshotRepo.SetAllInactiveByProjectID(project.ID, rolledBackBy)
	s.snapshotRepo.UpdateStatus(snapshotUUID, models.SnapshotStatusActive, rolledBackBy)

	return project.UUID, nil
}

// DeleteSnapshot soft deletes a snapshot
func (s *SnapshotService) DeleteSnapshot(snapshotUUID string, deletedBy string) error {
	return s.snapshotRepo.DeleteByUUID(snapshotUUID, deletedBy)
}

// GetSnapshotMetadata parses and returns the metadata from a snapshot
func (s *SnapshotService) GetSnapshotMetadata(snapshotUUID string) (*models.SnapshotMetadata, error) {
	snapshot, err := s.snapshotRepo.GetByUUID(snapshotUUID)
	if err != nil {
		return nil, fmt.Errorf("snapshot not found: %w", err)
	}

	var metadata models.SnapshotMetadata
	if err := json.Unmarshal(snapshot.Metadata, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &metadata, nil
}
