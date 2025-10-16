package repository

import (
	"database/sql"
	"fmt"
	"math/big"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/internal/models"
)

type ProjectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// uuidToInt64 converts UUID to int64 by taking first 8 bytes
func uuidToInt64(u uuid.UUID) int64 {
	bytes := u[:]
	// Take first 8 bytes and convert to int64
	var num big.Int
	num.SetBytes(bytes[:8])
	return num.Int64()
}

func (r *ProjectRepository) Create(project *models.Project) error {
	// Generate UUID v7
	uuidV7 := uuid.Must(uuid.NewV7())
	id := uuidToInt64(uuidV7)
	uuidStr := uuidV7.String()

	query := `
		INSERT INTO projects (id, uuid, name, description, status, namespace, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	_, err := r.db.Exec(query, id, uuidStr, project.Name, project.Description, project.Status, project.Namespace, project.CreatedBy)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	// Get the created project to populate all fields including timestamps
	createdProject, err := r.GetByUUID(uuidStr)
	if err != nil {
		return fmt.Errorf("failed to retrieve created project: %w", err)
	}

	*project = *createdProject
	return nil
}

// GetByUUID retrieves project by UUID (external identifier)
func (r *ProjectRepository) GetByUUID(uuid string) (*models.Project, error) {
	var project models.Project
	query := `
		SELECT id, uuid, name, description, status, namespace, git_repo_id,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM projects
		WHERE uuid = ? AND deleted_at IS NULL
	`

	err := r.db.Get(&project, query, uuid)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// GetByID retrieves project by internal ID (for internal use/FK joins)
func (r *ProjectRepository) GetByID(id int64) (*models.Project, error) {
	var project models.Project
	query := `
		SELECT id, uuid, name, description, status, namespace, git_repo_id,
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
		SELECT id, uuid, name, description, status, namespace, git_repo_id,
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
		WHERE uuid = ? AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query, project.Name, project.Description, project.Status, project.Namespace, project.UpdatedBy, project.UUID)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) UpdateStatusByUUID(uuid string, status string) error {
	query := `UPDATE projects SET status = ?, updated_at = NOW() WHERE uuid = ? AND deleted_at IS NULL`
	_, err := r.db.Exec(query, status, uuid)
	if err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}

	return nil
}

func (r *ProjectRepository) DeleteByUUID(uuid string, deletedBy string) error {
	// Soft delete
	query := `UPDATE projects SET deleted_by = ?, deleted_at = NOW() WHERE uuid = ? AND deleted_at IS NULL`
	_, err := r.db.Exec(query, deletedBy, uuid)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) GetWithGitRepoByUUID(uuid string) (*models.ProjectWithRelations, error) {
	project, err := r.GetByUUID(uuid)
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
		err = r.db.Get(&gitRepo, query, project.GitRepoID.Int64)
		if err == nil {
			result.GitRepo = &gitRepo
		}
	}

	return result, nil
}
