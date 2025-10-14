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
	// Validate project exists
	_, err := s.projectRepo.GetByID(req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Marshal fields to JSON
	fieldsJSON, err := json.Marshal(req.Fields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields: %w", err)
	}

	entity := &models.Entity{
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		TableName:   req.TableName,
		Fields:      fieldsJSON,
	}

	if req.Description != "" {
		entity.Description.String = req.Description
		entity.Description.Valid = true
	}

	err = s.repo.Create(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	return entity, nil
}

func (s *EntityService) GetEntity(id int64) (*models.Entity, error) {
	return s.repo.GetByID(id)
}

func (s *EntityService) GetEntitiesByProject(projectID int64) ([]models.Entity, error) {
	// Validate project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return s.repo.GetByProjectID(projectID)
}

func (s *EntityService) UpdateEntity(id int64, req *models.UpdateEntityRequest) (*models.Entity, error) {
	entity, err := s.repo.GetByID(id)
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

	err = s.repo.Update(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	return entity, nil
}

func (s *EntityService) DeleteEntity(id int64) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}
