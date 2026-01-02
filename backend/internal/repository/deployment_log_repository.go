package repository

import (
	"database/sql"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/internal/models"
)

// logIDCounter is used to generate unique log IDs
var logIDCounter int64

type DeploymentLogRepository struct {
	db *sqlx.DB
}

func NewDeploymentLogRepository(db *sqlx.DB) *DeploymentLogRepository {
	return &DeploymentLogRepository{db: db}
}

// generateLogID generates a unique ID for deployment logs
func generateLogID() int64 {
	return time.Now().UnixNano() + atomic.AddInt64(&logIDCounter, 1)
}

// Create creates a new deployment log entry
func (r *DeploymentLogRepository) Create(log *models.DeploymentLog) error {
	id := generateLogID()

	query := `
		INSERT INTO deployment_logs (id, deployment_id, level, message, step, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`

	var step sql.NullString
	if log.Step != "" {
		step = sql.NullString{String: log.Step, Valid: true}
	}

	_, err := r.db.Exec(query, id, log.DeploymentID, log.Level, log.Message, step)
	if err != nil {
		return fmt.Errorf("failed to create deployment log: %w", err)
	}

	log.ID = id
	log.CreatedAt = time.Now()
	return nil
}

// CreateBatch creates multiple log entries in a single transaction
func (r *DeploymentLogRepository) CreateBatch(logs []models.DeploymentLog) error {
	if len(logs) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	query := `
		INSERT INTO deployment_logs (id, deployment_id, level, message, step, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i := range logs {
		id := generateLogID()
		var step sql.NullString
		if logs[i].Step != "" {
			step = sql.NullString{String: logs[i].Step, Valid: true}
		}

		_, err := stmt.Exec(id, logs[i].DeploymentID, logs[i].Level, logs[i].Message, step)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert log: %w", err)
		}
		logs[i].ID = id
		logs[i].CreatedAt = time.Now()
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByDeploymentID retrieves all logs for a deployment
func (r *DeploymentLogRepository) GetByDeploymentID(deploymentID int64) ([]models.DeploymentLog, error) {
	var logs []models.DeploymentLog
	query := `
		SELECT id, deployment_id, level, message, COALESCE(step, '') as step, created_at
		FROM deployment_logs
		WHERE deployment_id = ?
		ORDER BY created_at ASC, id ASC
	`

	err := r.db.Select(&logs, query, deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment logs: %w", err)
	}

	return logs, nil
}

// GetByDeploymentIDPaginated retrieves logs with pagination
func (r *DeploymentLogRepository) GetByDeploymentIDPaginated(deploymentID int64, limit, offset int) ([]models.DeploymentLog, int64, error) {
	var logs []models.DeploymentLog
	query := `
		SELECT id, deployment_id, level, message, COALESCE(step, '') as step, created_at
		FROM deployment_logs
		WHERE deployment_id = ?
		ORDER BY created_at ASC, id ASC
		LIMIT ? OFFSET ?
	`

	err := r.db.Select(&logs, query, deploymentID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deployment logs: %w", err)
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM deployment_logs WHERE deployment_id = ?`
	err = r.db.Get(&total, countQuery, deploymentID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deployment logs: %w", err)
	}

	return logs, total, nil
}

// GetByDeploymentIDSince retrieves logs created after a specific time (for polling/streaming)
func (r *DeploymentLogRepository) GetByDeploymentIDSince(deploymentID int64, since time.Time) ([]models.DeploymentLog, error) {
	var logs []models.DeploymentLog
	query := `
		SELECT id, deployment_id, level, message, COALESCE(step, '') as step, created_at
		FROM deployment_logs
		WHERE deployment_id = ? AND created_at > ?
		ORDER BY created_at ASC, id ASC
	`

	err := r.db.Select(&logs, query, deploymentID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment logs since: %w", err)
	}

	return logs, nil
}

// GetByDeploymentIDAfterId retrieves logs with ID greater than specified (for SSE streaming)
func (r *DeploymentLogRepository) GetByDeploymentIDAfterId(deploymentID int64, afterID int64) ([]models.DeploymentLog, error) {
	var logs []models.DeploymentLog
	query := `
		SELECT id, deployment_id, level, message, COALESCE(step, '') as step, created_at
		FROM deployment_logs
		WHERE deployment_id = ? AND id > ?
		ORDER BY id ASC
	`

	err := r.db.Select(&logs, query, deploymentID, afterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment logs after id: %w", err)
	}

	return logs, nil
}

// GetLatestByDeploymentID retrieves the latest N logs for a deployment
func (r *DeploymentLogRepository) GetLatestByDeploymentID(deploymentID int64, limit int) ([]models.DeploymentLog, error) {
	var logs []models.DeploymentLog
	query := `
		SELECT id, deployment_id, level, message, COALESCE(step, '') as step, created_at
		FROM deployment_logs
		WHERE deployment_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`

	err := r.db.Select(&logs, query, deploymentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest deployment logs: %w", err)
	}

	// Reverse the order to show oldest first
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	return logs, nil
}

// GetByLevel retrieves logs by level for a deployment
func (r *DeploymentLogRepository) GetByLevel(deploymentID int64, level string) ([]models.DeploymentLog, error) {
	var logs []models.DeploymentLog
	query := `
		SELECT id, deployment_id, level, message, COALESCE(step, '') as step, created_at
		FROM deployment_logs
		WHERE deployment_id = ? AND level = ?
		ORDER BY created_at ASC, id ASC
	`

	err := r.db.Select(&logs, query, deploymentID, level)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment logs by level: %w", err)
	}

	return logs, nil
}

// GetErrorLogs retrieves only error logs for a deployment
func (r *DeploymentLogRepository) GetErrorLogs(deploymentID int64) ([]models.DeploymentLog, error) {
	return r.GetByLevel(deploymentID, models.LogLevelError)
}

// CountByDeploymentID counts total logs for a deployment
func (r *DeploymentLogRepository) CountByDeploymentID(deploymentID int64) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM deployment_logs WHERE deployment_id = ?`
	err := r.db.Get(&count, query, deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to count deployment logs: %w", err)
	}
	return count, nil
}

// DeleteByDeploymentID deletes all logs for a deployment
func (r *DeploymentLogRepository) DeleteByDeploymentID(deploymentID int64) error {
	query := `DELETE FROM deployment_logs WHERE deployment_id = ?`
	_, err := r.db.Exec(query, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to delete deployment logs: %w", err)
	}

	return nil
}

// LogInfo creates an info level log
func (r *DeploymentLogRepository) LogInfo(deploymentID int64, message string, step string) error {
	log := &models.DeploymentLog{
		DeploymentID: deploymentID,
		Level:        models.LogLevelInfo,
		Message:      message,
		Step:         step,
	}
	return r.Create(log)
}

// LogWarning creates a warning level log
func (r *DeploymentLogRepository) LogWarning(deploymentID int64, message string, step string) error {
	log := &models.DeploymentLog{
		DeploymentID: deploymentID,
		Level:        models.LogLevelWarning,
		Message:      message,
		Step:         step,
	}
	return r.Create(log)
}

// LogError creates an error level log
func (r *DeploymentLogRepository) LogError(deploymentID int64, message string, step string) error {
	log := &models.DeploymentLog{
		DeploymentID: deploymentID,
		Level:        models.LogLevelError,
		Message:      message,
		Step:         step,
	}
	return r.Create(log)
}

// LogDebug creates a debug level log
func (r *DeploymentLogRepository) LogDebug(deploymentID int64, message string, step string) error {
	log := &models.DeploymentLog{
		DeploymentID: deploymentID,
		Level:        models.LogLevelDebug,
		Message:      message,
		Step:         step,
	}
	return r.Create(log)
}
