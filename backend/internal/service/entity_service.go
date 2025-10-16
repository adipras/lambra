package service

import (
	"encoding/json"
	"fmt"

	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
)

type EntityService struct {
	repo        *repository.EntityRepository
	projectRepo *repository.ProjectRepository
}

func NewEntityService(repo *repository.EntityRepository, projectRepo *repository.ProjectRepository) *EntityService {
	return &EntityService{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

func (s *EntityService) CreateEntity(req *models.CreateEntityRequest) (*models.Entity, error) {
	// Validate project exists and get internal ID
	project, err := s.projectRepo.GetByUUID(req.ProjectUUID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Marshal fields to JSON
	fieldsJSON, err := json.Marshal(req.Fields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields: %w", err)
	}

	entity := &models.Entity{
		ProjectID: project.ID, // Use internal project ID
		Name:      req.Name,
		TableName: req.TableName,
		Fields:    fieldsJSON,
	}

	if req.Description != "" {
		entity.Description.String = req.Description
		entity.Description.Valid = true
	}

	// Set created_by (in future, get from auth context)
	entity.SetCreatedBy("system")

	err = s.repo.Create(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	return entity, nil
}

func (s *EntityService) GetEntityByUUID(uuid string) (*models.Entity, error) {
	return s.repo.GetByUUID(uuid)
}

func (s *EntityService) GetEntitiesByProjectUUID(projectUUID string) ([]models.Entity, error) {
	// Validate project exists and get internal ID
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return s.repo.GetByProjectID(project.ID)
}

func (s *EntityService) UpdateEntity(uuid string, req *models.UpdateEntityRequest) (*models.Entity, error) {
	entity, err := s.repo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		entity.Name = req.Name
	}
	if req.TableName != "" {
		entity.TableName = req.TableName
	}
	if req.Description != "" {
		entity.Description.String = req.Description
		entity.Description.Valid = true
	}
	if len(req.Fields) > 0 {
		fieldsJSON, err := json.Marshal(req.Fields)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal fields: %w", err)
		}
		entity.Fields = fieldsJSON
	}

	// Set updated_by (in future, get from auth context)
	entity.SetUpdatedBy("system")

	err = s.repo.Update(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	return entity, nil
}

func (s *EntityService) DeleteEntity(uuid string) error {
	_, err := s.repo.GetByUUID(uuid)
	if err != nil {
		return err
	}

	// Soft delete with deleted_by (in future, get from auth context)
	return s.repo.DeleteByUUID(uuid, "system")
}
