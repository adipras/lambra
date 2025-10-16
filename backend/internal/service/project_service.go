package service

import (
	"database/sql"
	"fmt"

	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
)

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) CreateProject(req *models.CreateProjectRequest) (*models.Project, error) {
	project := &models.Project{
		Name:      req.Name,
		Namespace: req.Namespace,
		Status:    models.ProjectStatusActive,
	}

	if req.Description != "" {
		project.Description = sql.NullString{String: req.Description, Valid: true}
	}

	// Set created_by (in future, get from auth context)
	project.SetCreatedBy("system")

	err := s.repo.Create(project)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

func (s *ProjectService) GetProjectByUUID(uuid string) (*models.Project, error) {
	return s.repo.GetByUUID(uuid)
}

func (s *ProjectService) GetProjectWithRelations(uuid string) (*models.ProjectWithRelations, error) {
	return s.repo.GetWithGitRepoByUUID(uuid)
}

func (s *ProjectService) GetAllProjects(page, limit int) ([]models.Project, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit
	return s.repo.GetAll(limit, offset)
}

func (s *ProjectService) UpdateProject(uuid string, req *models.UpdateProjectRequest) (*models.Project, error) {
	project, err := s.repo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.Status != "" {
		project.Status = req.Status
	}

	// Set updated_by (in future, get from auth context)
	project.SetUpdatedBy("system")

	err = s.repo.Update(project)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

func (s *ProjectService) DeleteProject(uuid string) error {
	_, err := s.repo.GetByUUID(uuid)
	if err != nil {
		return err
	}

	// Soft delete with deleted_by (in future, get from auth context)
	return s.repo.DeleteByUUID(uuid, "system")
}
