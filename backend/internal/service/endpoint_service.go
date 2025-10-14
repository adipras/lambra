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
	// Validate entity exists
	_, err := s.entityRepo.GetByID(req.EntityID)
	if err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}

	// Validate project exists
	_, err = s.projectRepo.GetByID(req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	endpoint := &models.Endpoint{
		EntityID:       req.EntityID,
		ProjectID:      req.ProjectID,
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

	err = s.repo.Create(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	return endpoint, nil
}

func (s *EndpointService) GetEndpoint(id int64) (*models.Endpoint, error) {
	return s.repo.GetByID(id)
}

func (s *EndpointService) GetEndpointsByProject(projectID int64) ([]models.Endpoint, error) {
	// Validate project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return s.repo.GetByProjectID(projectID)
}

func (s *EndpointService) GetEndpointsByEntity(entityID int64) ([]models.Endpoint, error) {
	// Validate entity exists
	_, err := s.entityRepo.GetByID(entityID)
	if err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}

	return s.repo.GetByEntityID(entityID)
}

func (s *EndpointService) UpdateEndpoint(id int64, req *models.UpdateEndpointRequest) (*models.Endpoint, error) {
	endpoint, err := s.repo.GetByID(id)
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

	err = s.repo.Update(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to update endpoint: %w", err)
	}

	return endpoint, nil
}

func (s *EndpointService) DeleteEndpoint(id int64) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}
