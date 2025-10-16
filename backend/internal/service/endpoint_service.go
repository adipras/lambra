package service

import (
	"fmt"

	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
)

type EndpointService struct {
	repo        *repository.EndpointRepository
	entityRepo  *repository.EntityRepository
	projectRepo *repository.ProjectRepository
}

func NewEndpointService(repo *repository.EndpointRepository, entityRepo *repository.EntityRepository, projectRepo *repository.ProjectRepository) *EndpointService {
	return &EndpointService{
		repo:        repo,
		entityRepo:  entityRepo,
		projectRepo: projectRepo,
	}
}

func (s *EndpointService) CreateEndpoint(req *models.CreateEndpointRequest) (*models.Endpoint, error) {
	// Validate entity exists and get internal ID
	entity, err := s.entityRepo.GetByUUID(req.EntityUUID)
	if err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}

	endpoint := &models.Endpoint{
		EntityID:       entity.ID,      // Use internal entity ID
		ProjectID:      entity.ProjectID, // Get project ID from entity
		Name:           req.Name,
		Path:           req.Path,
		Method:         req.Method,
		RequestSchema:  req.RequestSchema,
		ResponseSchema: req.ResponseSchema,
		RequireAuth:    req.RequireAuth,
	}

	if req.Description != "" {
		endpoint.Description.String = req.Description
		endpoint.Description.Valid = true
	}

	// Set created_by (in future, get from auth context)
	endpoint.SetCreatedBy("system")

	err = s.repo.Create(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	return endpoint, nil
}

func (s *EndpointService) GetEndpointByUUID(uuid string) (*models.Endpoint, error) {
	return s.repo.GetByUUID(uuid)
}

func (s *EndpointService) GetEndpointsByProjectUUID(projectUUID string) ([]models.Endpoint, error) {
	// Validate project exists and get internal ID
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return s.repo.GetByProjectID(project.ID)
}

func (s *EndpointService) GetEndpointsByEntityUUID(entityUUID string) ([]models.Endpoint, error) {
	// Validate entity exists and get internal ID
	entity, err := s.entityRepo.GetByUUID(entityUUID)
	if err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}

	return s.repo.GetByEntityID(entity.ID)
}

func (s *EndpointService) UpdateEndpoint(uuid string, req *models.UpdateEndpointRequest) (*models.Endpoint, error) {
	endpoint, err := s.repo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		endpoint.Name = req.Name
	}
	if req.Path != "" {
		endpoint.Path = req.Path
	}
	if req.Method != "" {
		endpoint.Method = req.Method
	}
	if req.Description != "" {
		endpoint.Description.String = req.Description
		endpoint.Description.Valid = true
	}
	if req.RequestSchema != nil {
		endpoint.RequestSchema = req.RequestSchema
	}
	if req.ResponseSchema != nil {
		endpoint.ResponseSchema = req.ResponseSchema
	}
	if req.RequireAuth != nil {
		endpoint.RequireAuth = *req.RequireAuth
	}

	// Set updated_by (in future, get from auth context)
	endpoint.SetUpdatedBy("system")

	err = s.repo.Update(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to update endpoint: %w", err)
	}

	return endpoint, nil
}

func (s *EndpointService) DeleteEndpoint(uuid string) error {
	_, err := s.repo.GetByUUID(uuid)
	if err != nil {
		return err
	}

	// Soft delete with deleted_by (in future, get from auth context)
	return s.repo.DeleteByUUID(uuid, "system")
}
