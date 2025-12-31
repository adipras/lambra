package repository

import (
	"database/sql"
	"fmt"
	"math/big"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/internal/models"
)

type SnapshotRepository struct {
	db *sqlx.DB
}

func NewSnapshotRepository(db *sqlx.DB) *SnapshotRepository {
	return &SnapshotRepository{db: db}
}

// uuidToInt64Snapshot converts UUID to int64 by taking first 8 bytes
func uuidToInt64Snapshot(u uuid.UUID) int64 {
	bytes := u[:]
	var num big.Int
	num.SetBytes(bytes[:8])
	return num.Int64()
}

func (r *SnapshotRepository) Create(snapshot *models.GenerationSnapshot) error {
	// Generate UUID v7
	uuidV7 := uuid.Must(uuid.NewV7())
	id := uuidToInt64Snapshot(uuidV7)
	uuidStr := uuidV7.String()

	query := `
		INSERT INTO generation_snapshots (
			id, uuid, project_id, version, git_commit_hash, git_tag,
			metadata, database_snapshot, status, created_by, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := r.db.Exec(query,
		id, uuidStr, snapshot.ProjectID, snapshot.Version, snapshot.GitCommitHash,
		snapshot.GitTag, snapshot.Metadata, snapshot.DatabaseSnapshot,
		snapshot.Status, snapshot.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Get the created snapshot to populate all fields
	createdSnapshot, err := r.GetByUUID(uuidStr)
	if err != nil {
		return fmt.Errorf("failed to retrieve created snapshot: %w", err)
	}

	*snapshot = *createdSnapshot
	return nil
}

// GetByUUID retrieves snapshot by UUID (external identifier)
func (r *SnapshotRepository) GetByUUID(uuid string) (*models.GenerationSnapshot, error) {
	var snapshot models.GenerationSnapshot
	query := `
		SELECT id, uuid, project_id, version, git_commit_hash, git_tag,
		       metadata, database_snapshot, status,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM generation_snapshots
		WHERE uuid = ? AND deleted_at IS NULL
	`

	err := r.db.Get(&snapshot, query, uuid)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("snapshot not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	return &snapshot, nil
}

// GetByID retrieves snapshot by internal ID (for FK joins)
func (r *SnapshotRepository) GetByID(id int64) (*models.GenerationSnapshot, error) {
	var snapshot models.GenerationSnapshot
	query := `
		SELECT id, uuid, project_id, version, git_commit_hash, git_tag,
		       metadata, database_snapshot, status,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM generation_snapshots
		WHERE id = ? AND deleted_at IS NULL
	`

	err := r.db.Get(&snapshot, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("snapshot not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	return &snapshot, nil
}

// GetByProjectID retrieves all snapshots for a project with pagination
func (r *SnapshotRepository) GetByProjectID(projectID int64, limit, offset int) ([]models.GenerationSnapshot, int64, error) {
	var snapshots []models.GenerationSnapshot
	query := `
		SELECT id, uuid, project_id, version, git_commit_hash, git_tag,
		       metadata, database_snapshot, status,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM generation_snapshots
		WHERE project_id = ? AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	err := r.db.Select(&snapshots, query, projectID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get snapshots: %w", err)
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM generation_snapshots WHERE project_id = ? AND deleted_at IS NULL`
	err = r.db.Get(&total, countQuery, projectID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count snapshots: %w", err)
	}

	return snapshots, total, nil
}

// GetLatestByProjectID retrieves the latest snapshot for a project
func (r *SnapshotRepository) GetLatestByProjectID(projectID int64) (*models.GenerationSnapshot, error) {
	var snapshot models.GenerationSnapshot
	query := `
		SELECT id, uuid, project_id, version, git_commit_hash, git_tag,
		       metadata, database_snapshot, status,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM generation_snapshots
		WHERE project_id = ? AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.db.Get(&snapshot, query, projectID)
	if err == sql.ErrNoRows {
		return nil, nil // No snapshot found, not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest snapshot: %w", err)
	}

	return &snapshot, nil
}

// GetActiveByProjectID retrieves the active snapshot for a project
func (r *SnapshotRepository) GetActiveByProjectID(projectID int64) (*models.GenerationSnapshot, error) {
	var snapshot models.GenerationSnapshot
	query := `
		SELECT id, uuid, project_id, version, git_commit_hash, git_tag,
		       metadata, database_snapshot, status,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM generation_snapshots
		WHERE project_id = ? AND status = ? AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.db.Get(&snapshot, query, projectID, models.SnapshotStatusActive)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active snapshot: %w", err)
	}

	return &snapshot, nil
}

// UpdateStatus updates the status of a snapshot
func (r *SnapshotRepository) UpdateStatus(uuid string, status string, updatedBy string) error {
	query := `
		UPDATE generation_snapshots
		SET status = ?, updated_by = ?, updated_at = NOW()
		WHERE uuid = ? AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query, status, updatedBy, uuid)
	if err != nil {
		return fmt.Errorf("failed to update snapshot status: %w", err)
	}

	return nil
}

// SetAllInactiveByProjectID sets all snapshots for a project to rolled_back status
// (used before activating a new snapshot)
func (r *SnapshotRepository) SetAllInactiveByProjectID(projectID int64, updatedBy string) error {
	query := `
		UPDATE generation_snapshots
		SET status = ?, updated_by = ?, updated_at = NOW()
		WHERE project_id = ? AND status = ? AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query, models.SnapshotStatusRolledBack, updatedBy, projectID, models.SnapshotStatusActive)
	if err != nil {
		return fmt.Errorf("failed to deactivate snapshots: %w", err)
	}

	return nil
}

// DeleteByUUID soft deletes a snapshot
func (r *SnapshotRepository) DeleteByUUID(uuid string, deletedBy string) error {
	query := `
		UPDATE generation_snapshots
		SET deleted_by = ?, deleted_at = NOW()
		WHERE uuid = ? AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query, deletedBy, uuid)
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return nil
}

// HardDeleteByUUID permanently deletes a snapshot
func (r *SnapshotRepository) HardDeleteByUUID(uuid string) error {
	query := `DELETE FROM generation_snapshots WHERE uuid = ?`
	_, err := r.db.Exec(query, uuid)
	if err != nil {
		return fmt.Errorf("failed to hard delete snapshot: %w", err)
	}

	return nil
}

// HardDeleteByProjectID permanently deletes all snapshots for a project
func (r *SnapshotRepository) HardDeleteByProjectID(projectID int64) error {
	query := `DELETE FROM generation_snapshots WHERE project_id = ?`
	_, err := r.db.Exec(query, projectID)
	if err != nil {
		return fmt.Errorf("failed to hard delete snapshots by project: %w", err)
	}

	return nil
}

// CountByProjectID counts total snapshots for a project
func (r *SnapshotRepository) CountByProjectID(projectID int64) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM generation_snapshots WHERE project_id = ? AND deleted_at IS NULL`
	err := r.db.Get(&count, query, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to count snapshots: %w", err)
	}
	return count, nil
}
