package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
)

type SnapshotService struct {
	snapshotRepo *repository.SnapshotRepository
	projectRepo  *repository.ProjectRepository
	entityRepo   *repository.EntityRepository
	endpointRepo *repository.EndpointRepository
}

func NewSnapshotService(
	snapshotRepo *repository.SnapshotRepository,
	projectRepo *repository.ProjectRepository,
	entityRepo *repository.EntityRepository,
	endpointRepo *repository.EndpointRepository,
) *SnapshotService {
	return &SnapshotService{
		snapshotRepo: snapshotRepo,
		projectRepo:  projectRepo,
		entityRepo:   entityRepo,
		endpointRepo: endpointRepo,
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

	// Get all endpoints for each entity
	var allEndpoints []models.Endpoint
	for _, entity := range entities {
		endpoints, err := s.endpointRepo.GetByEntityID(entity.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get endpoints for entity %s: %w", entity.Name, err)
		}
		allEndpoints = append(allEndpoints, endpoints...)
	}

	// Create snapshot metadata
	metadata := models.SnapshotMetadata{
		Entities:  entities,
		Endpoints: allEndpoints,
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

	// Soft delete all current entities (this will cascade to endpoints due to FK)
	currentEntities, err := s.entityRepo.GetByProjectID(project.ID)
	if err == nil {
		for _, entity := range currentEntities {
			s.entityRepo.DeleteByUUID(entity.UUID, rolledBackBy)
		}
	}

	// Restore entities from snapshot
	entityIDMap := make(map[int64]int64) // old ID -> new ID
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

		entityIDMap[entity.ID] = newEntity.ID
	}

	// Restore endpoints from snapshot
	for _, endpoint := range metadata.Endpoints {
		newEntityID, ok := entityIDMap[endpoint.EntityID]
		if !ok {
			continue // Skip if entity wasn't restored
		}

		newEndpoint := &models.Endpoint{
			EntityID:        newEntityID,
			ProjectID:       project.ID,
			Name:            endpoint.Name,
			Path:            endpoint.Path,
			Method:          endpoint.Method,
			Description:     endpoint.Description,
			RequestSchema:   endpoint.RequestSchema,
			ResponseSchema:  endpoint.ResponseSchema,
			RequireAuth:     endpoint.RequireAuth,
		}
		newEndpoint.SetCreatedBy(rolledBackBy)

		if err := s.endpointRepo.Create(newEndpoint); err != nil {
			fmt.Printf("Warning: failed to restore endpoint %s: %v\n", endpoint.Name, err)
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
