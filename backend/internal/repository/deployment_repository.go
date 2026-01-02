package repository

import (
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/internal/models"
)

type DeploymentRepository struct {
	db *sqlx.DB
}

func NewDeploymentRepository(db *sqlx.DB) *DeploymentRepository {
	return &DeploymentRepository{db: db}
}

// uuidToInt64Deployment converts UUID to int64 by taking first 8 bytes
func uuidToInt64Deployment(u uuid.UUID) int64 {
	bytes := u[:]
	var num big.Int
	num.SetBytes(bytes[:8])
	return num.Int64()
}

// Create creates a new deployment record
func (r *DeploymentRepository) Create(deployment *models.Deployment) error {
	// Generate UUID v7
	uuidV7 := uuid.Must(uuid.NewV7())
	id := uuidToInt64Deployment(uuidV7)
	uuidStr := uuidV7.String()

	query := `
		INSERT INTO deployments (
			id, uuid, project_id, snapshot_id, environment, status, version,
			deployed_by, deployment_url, error_message, started_at,
			created_by, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := r.db.Exec(query,
		id, uuidStr, deployment.ProjectID, deployment.SnapshotID,
		deployment.Environment, deployment.Status, deployment.Version,
		deployment.DeployedBy, deployment.DeploymentURL, deployment.ErrorMessage,
		deployment.StartedAt, deployment.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	// Get the created deployment to populate all fields
	createdDeployment, err := r.GetByUUID(uuidStr)
	if err != nil {
		return fmt.Errorf("failed to retrieve created deployment: %w", err)
	}

	*deployment = *createdDeployment
	return nil
}

// GetByUUID retrieves deployment by UUID (external identifier)
func (r *DeploymentRepository) GetByUUID(uuid string) (*models.Deployment, error) {
	var deployment models.Deployment
	query := `
		SELECT d.id, d.uuid, d.project_id, d.snapshot_id, d.environment, d.status,
		       d.version, d.deployed_by, d.deployment_url, d.error_message,
		       d.started_at, d.completed_at,
		       d.created_by, d.updated_by, d.deleted_by, d.created_at, d.updated_at, d.deleted_at,
		       p.uuid as project_uuid
		FROM deployments d
		LEFT JOIN projects p ON d.project_id = p.id
		WHERE d.uuid = ? AND d.deleted_at IS NULL
	`

	row := r.db.QueryRowx(query, uuid)
	var projectUUID sql.NullString
	err := row.Scan(
		&deployment.ID, &deployment.UUID, &deployment.ProjectID, &deployment.SnapshotID,
		&deployment.Environment, &deployment.Status, &deployment.Version,
		&deployment.DeployedBy, &deployment.DeploymentURL, &deployment.ErrorMessage,
		&deployment.StartedAt, &deployment.CompletedAt,
		&deployment.CreatedBy, &deployment.UpdatedBy, &deployment.DeletedBy,
		&deployment.CreatedAt, &deployment.UpdatedAt, &deployment.DeletedAt,
		&projectUUID,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("deployment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	if projectUUID.Valid {
		deployment.ProjectUUID = projectUUID.String
	}

	return &deployment, nil
}

// GetByID retrieves deployment by internal ID (for FK joins)
func (r *DeploymentRepository) GetByID(id int64) (*models.Deployment, error) {
	var deployment models.Deployment
	query := `
		SELECT d.id, d.uuid, d.project_id, d.snapshot_id, d.environment, d.status,
		       d.version, d.deployed_by, d.deployment_url, d.error_message,
		       d.started_at, d.completed_at,
		       d.created_by, d.updated_by, d.deleted_by, d.created_at, d.updated_at, d.deleted_at,
		       p.uuid as project_uuid
		FROM deployments d
		LEFT JOIN projects p ON d.project_id = p.id
		WHERE d.id = ? AND d.deleted_at IS NULL
	`

	row := r.db.QueryRowx(query, id)
	var projectUUID sql.NullString
	err := row.Scan(
		&deployment.ID, &deployment.UUID, &deployment.ProjectID, &deployment.SnapshotID,
		&deployment.Environment, &deployment.Status, &deployment.Version,
		&deployment.DeployedBy, &deployment.DeploymentURL, &deployment.ErrorMessage,
		&deployment.StartedAt, &deployment.CompletedAt,
		&deployment.CreatedBy, &deployment.UpdatedBy, &deployment.DeletedBy,
		&deployment.CreatedAt, &deployment.UpdatedAt, &deployment.DeletedAt,
		&projectUUID,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("deployment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	if projectUUID.Valid {
		deployment.ProjectUUID = projectUUID.String
	}

	return &deployment, nil
}

// GetByProjectID retrieves all deployments for a project with pagination
func (r *DeploymentRepository) GetByProjectID(projectID int64, limit, offset int) ([]models.Deployment, int64, error) {
	query := `
		SELECT d.id, d.uuid, d.project_id, d.snapshot_id, d.environment, d.status,
		       d.version, d.deployed_by, d.deployment_url, d.error_message,
		       d.started_at, d.completed_at,
		       d.created_by, d.updated_by, d.deleted_by, d.created_at, d.updated_at, d.deleted_at,
		       p.uuid as project_uuid
		FROM deployments d
		LEFT JOIN projects p ON d.project_id = p.id
		WHERE d.project_id = ? AND d.deleted_at IS NULL
		ORDER BY d.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Queryx(query, projectID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deployments: %w", err)
	}
	defer rows.Close()

	var deployments []models.Deployment
	for rows.Next() {
		var deployment models.Deployment
		var projectUUID sql.NullString
		err := rows.Scan(
			&deployment.ID, &deployment.UUID, &deployment.ProjectID, &deployment.SnapshotID,
			&deployment.Environment, &deployment.Status, &deployment.Version,
			&deployment.DeployedBy, &deployment.DeploymentURL, &deployment.ErrorMessage,
			&deployment.StartedAt, &deployment.CompletedAt,
			&deployment.CreatedBy, &deployment.UpdatedBy, &deployment.DeletedBy,
			&deployment.CreatedAt, &deployment.UpdatedAt, &deployment.DeletedAt,
			&projectUUID,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan deployment: %w", err)
		}
		if projectUUID.Valid {
			deployment.ProjectUUID = projectUUID.String
		}
		deployments = append(deployments, deployment)
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM deployments WHERE project_id = ? AND deleted_at IS NULL`
	err = r.db.Get(&total, countQuery, projectID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deployments: %w", err)
	}

	return deployments, total, nil
}

// GetLatestByProjectID retrieves the latest deployment for a project
func (r *DeploymentRepository) GetLatestByProjectID(projectID int64) (*models.Deployment, error) {
	deployments, _, err := r.GetByProjectID(projectID, 1, 0)
	if err != nil {
		return nil, err
	}
	if len(deployments) == 0 {
		return nil, nil // No deployment found, not an error
	}
	return &deployments[0], nil
}

// UpdateStatus updates the status of a deployment
func (r *DeploymentRepository) UpdateStatus(uuid string, status string, errorMessage string) error {
	query := `
		UPDATE deployments
		SET status = ?, error_message = ?, updated_at = NOW()
		WHERE uuid = ? AND deleted_at IS NULL
	`
	var errMsg sql.NullString
	if errorMessage != "" {
		errMsg = sql.NullString{String: errorMessage, Valid: true}
	}

	_, err := r.db.Exec(query, status, errMsg, uuid)
	if err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	return nil
}

// SetCompleted marks a deployment as completed
func (r *DeploymentRepository) SetCompleted(uuid string, status string, deploymentURL string) error {
	query := `
		UPDATE deployments
		SET status = ?, deployment_url = ?, completed_at = NOW(), updated_at = NOW()
		WHERE uuid = ? AND deleted_at IS NULL
	`
	var url sql.NullString
	if deploymentURL != "" {
		url = sql.NullString{String: deploymentURL, Valid: true}
	}

	_, err := r.db.Exec(query, status, url, uuid)
	if err != nil {
		return fmt.Errorf("failed to complete deployment: %w", err)
	}

	return nil
}

// SetFailed marks a deployment as failed
func (r *DeploymentRepository) SetFailed(uuid string, errorMessage string) error {
	query := `
		UPDATE deployments
		SET status = ?, error_message = ?, completed_at = NOW(), updated_at = NOW()
		WHERE uuid = ? AND deleted_at IS NULL
	`

	_, err := r.db.Exec(query, models.DeploymentStatusFailed, errorMessage, uuid)
	if err != nil {
		return fmt.Errorf("failed to mark deployment as failed: %w", err)
	}

	return nil
}

// DeleteByUUID soft deletes a deployment
func (r *DeploymentRepository) DeleteByUUID(uuid string, deletedBy string) error {
	query := `
		UPDATE deployments
		SET deleted_by = ?, deleted_at = NOW()
		WHERE uuid = ? AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query, deletedBy, uuid)
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	return nil
}

// HardDeleteByUUID permanently deletes a deployment
func (r *DeploymentRepository) HardDeleteByUUID(uuid string) error {
	query := `DELETE FROM deployments WHERE uuid = ?`
	_, err := r.db.Exec(query, uuid)
	if err != nil {
		return fmt.Errorf("failed to hard delete deployment: %w", err)
	}

	return nil
}

// HardDeleteByProjectID permanently deletes all deployments for a project
func (r *DeploymentRepository) HardDeleteByProjectID(projectID int64) error {
	query := `DELETE FROM deployments WHERE project_id = ?`
	_, err := r.db.Exec(query, projectID)
	if err != nil {
		return fmt.Errorf("failed to hard delete deployments by project: %w", err)
	}

	return nil
}

// CountByProjectID counts total deployments for a project
func (r *DeploymentRepository) CountByProjectID(projectID int64) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM deployments WHERE project_id = ? AND deleted_at IS NULL`
	err := r.db.Get(&count, query, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to count deployments: %w", err)
	}
	return count, nil
}

// GetActiveByProjectID retrieves the currently active (deploying or success) deployment for a project
func (r *DeploymentRepository) GetActiveByProjectID(projectID int64) (*models.Deployment, error) {
	query := `
		SELECT d.id, d.uuid, d.project_id, d.snapshot_id, d.environment, d.status,
		       d.version, d.deployed_by, d.deployment_url, d.error_message,
		       d.started_at, d.completed_at,
		       d.created_by, d.updated_by, d.deleted_by, d.created_at, d.updated_at, d.deleted_at,
		       p.uuid as project_uuid
		FROM deployments d
		LEFT JOIN projects p ON d.project_id = p.id
		WHERE d.project_id = ? AND d.status IN (?, ?) AND d.deleted_at IS NULL
		ORDER BY d.created_at DESC
		LIMIT 1
	`

	row := r.db.QueryRowx(query, projectID, models.DeploymentStatusDeploying, models.DeploymentStatusSuccess)
	var deployment models.Deployment
	var projectUUID sql.NullString
	err := row.Scan(
		&deployment.ID, &deployment.UUID, &deployment.ProjectID, &deployment.SnapshotID,
		&deployment.Environment, &deployment.Status, &deployment.Version,
		&deployment.DeployedBy, &deployment.DeploymentURL, &deployment.ErrorMessage,
		&deployment.StartedAt, &deployment.CompletedAt,
		&deployment.CreatedBy, &deployment.UpdatedBy, &deployment.DeletedBy,
		&deployment.CreatedAt, &deployment.UpdatedAt, &deployment.DeletedAt,
		&projectUUID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active deployment: %w", err)
	}

	if projectUUID.Valid {
		deployment.ProjectUUID = projectUUID.String
	}

	return &deployment, nil
}

// GetDeploymentDuration returns the duration of a completed deployment
func (r *DeploymentRepository) GetDeploymentDuration(uuid string) (time.Duration, error) {
	deployment, err := r.GetByUUID(uuid)
	if err != nil {
		return 0, err
	}

	if !deployment.CompletedAt.Valid {
		return time.Since(deployment.StartedAt), nil
	}

	return deployment.CompletedAt.Time.Sub(deployment.StartedAt), nil
}
