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
		Name:        req.Name,
		Namespace:   req.Namespace,
		Status:      models.ProjectStatusActive,
	}

	if req.Description != "" {
		project.Description = sql.NullString{String: req.Description, Valid: true}
	}

	err := s.repo.Create(project)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

func (s *ProjectService) GetProject(id int64) (*models.Project, error) {
	return s.repo.GetByID(id)
}

func (s *ProjectService) GetProjectWithRelations(id int64) (*models.ProjectWithRelations, error) {
	return s.repo.GetWithGitRepo(id)
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

func (s *ProjectService) UpdateProject(id int64, req *models.UpdateProjectRequest) (*models.Project, error) {
	project, err := s.repo.GetByID(id)
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

	err = s.repo.Update(project)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

func (s *ProjectService) DeleteProject(id int64) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}
