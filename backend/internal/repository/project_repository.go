package repository

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/internal/models"
)

type ProjectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(project *models.Project) error {
	// Generate UUID using database function
	var generatedID string
	err := r.db.Get(&generatedID, "SELECT UUID()")
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}

	query := `
		INSERT INTO projects (id, name, description, status, namespace, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	_, err = r.db.Exec(query, generatedID, project.Name, project.Description, project.Status, project.Namespace, project.CreatedBy)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	// Get the created project to populate all fields including timestamps
	createdProject, err := r.GetByID(generatedID)
	if err != nil {
		return fmt.Errorf("failed to retrieve created project: %w", err)
	}

	*project = *createdProject
	return nil
}

func (r *ProjectRepository) GetByID(id string) (*models.Project, error) {
	var project models.Project
	query := `
		SELECT id, name, description, status, namespace, git_repo_id,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM projects
		WHERE id = ? AND deleted_at IS NULL
	`

	err := r.db.Get(&project, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

func (r *ProjectRepository) GetAll(limit, offset int) ([]models.Project, int64, error) {
	var projects []models.Project
	query := `
		SELECT id, name, description, status, namespace, git_repo_id,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM projects
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	err := r.db.Select(&projects, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get projects: %w", err)
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM projects WHERE deleted_at IS NULL`
	err = r.db.Get(&total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	return projects, total, nil
}

func (r *ProjectRepository) Update(project *models.Project) error {
	query := `
		UPDATE projects
		SET name = ?, description = ?, status = ?, namespace = ?, updated_by = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query, project.Name, project.Description, project.Status, project.Namespace, project.UpdatedBy, project.ID)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) UpdateStatus(id string, status string) error {
	query := `UPDATE projects SET status = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}

	return nil
}

func (r *ProjectRepository) Delete(id string, deletedBy string) error {
	// Soft delete
	query := `UPDATE projects SET deleted_by = ?, deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db.Exec(query, deletedBy, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) GetWithGitRepo(id string) (*models.ProjectWithRelations, error) {
	project, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	result := &models.ProjectWithRelations{
		Project: *project,
	}

	// Get Git repository if exists
	if project.GitRepoID.Valid {
		var gitRepo models.GitRepository
		query := `SELECT * FROM git_repositories WHERE id = ? AND deleted_at IS NULL`
		err = r.db.Get(&gitRepo, query, project.GitRepoID.String)
		if err == nil {
			result.GitRepo = &gitRepo
		}
	}

	return result, nil
}
