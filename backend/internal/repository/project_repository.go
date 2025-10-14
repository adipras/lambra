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
	query := `
		INSERT INTO projects (name, description, status, namespace, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`
	result, err := r.db.Exec(query, project.Name, project.Description, project.Status, project.Namespace)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	project.ID = id
	return nil
}

func (r *ProjectRepository) GetByID(id int64) (*models.Project, error) {
	var project models.Project
	query := `SELECT * FROM projects WHERE id = ?`

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
	query := `SELECT * FROM projects ORDER BY created_at DESC LIMIT ? OFFSET ?`

	err := r.db.Select(&projects, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get projects: %w", err)
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM projects`
	err = r.db.Get(&total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	return projects, total, nil
}

func (r *ProjectRepository) Update(project *models.Project) error {
	query := `
		UPDATE projects
		SET name = ?, description = ?, status = ?, namespace = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err := r.db.Exec(query, project.Name, project.Description, project.Status, project.Namespace, project.ID)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) UpdateStatus(id int64, status string) error {
	query := `UPDATE projects SET status = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}

	return nil
}

func (r *ProjectRepository) Delete(id int64) error {
	query := `DELETE FROM projects WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) GetWithGitRepo(id int64) (*models.ProjectWithRelations, error) {
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
		query := `SELECT * FROM git_repositories WHERE id = ?`
		err = r.db.Get(&gitRepo, query, project.GitRepoID.Int64)
		if err == nil {
			result.GitRepo = &gitRepo
		}
	}

	return result, nil
}
