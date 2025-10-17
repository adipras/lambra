package models

import (
	"database/sql"
	"time"
)

// Deployment represents a deployment instance
type Deployment struct {
	ID            int64          `db:"id" json:"id"`
	ProjectID     int64          `db:"project_id" json:"project_id"`
	SnapshotID    sql.NullInt64  `db:"snapshot_id" json:"snapshot_id,omitempty"`
	Environment   string         `db:"environment" json:"environment"` // dev, staging, production
	Status        string         `db:"status" json:"status"`           // pending, deploying, success, failed
	Version       string         `db:"version" json:"version"`
	DeployedBy    sql.NullString `db:"deployed_by" json:"deployed_by,omitempty"`
	DeploymentURL sql.NullString `db:"deployment_url" json:"deployment_url,omitempty"`
	ErrorMessage  sql.NullString `db:"error_message" json:"error_message,omitempty"`
	StartedAt     time.Time      `db:"started_at" json:"started_at"`
	CompletedAt   sql.NullTime   `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
}

// DeploymentWithLogs includes deployment logs
type DeploymentWithLogs struct {
	Deployment
	Logs []DeploymentLog `json:"logs,omitempty"`
}

// DeploymentLog represents deployment log entries
type DeploymentLog struct {
	ID           int64     `db:"id" json:"id"`
	DeploymentID int64     `db:"deployment_id" json:"deployment_id"`
	Level        string    `db:"level" json:"level"` // info, warning, error
	Message      string    `db:"message" json:"message"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// CreateDeploymentRequest for creating deployment
type CreateDeploymentRequest struct {
	ProjectID   int64  `json:"project_id" binding:"required"`
	SnapshotID  int64  `json:"snapshot_id"`
	Environment string `json:"environment" binding:"required,oneof=dev staging production"`
	Version     string `json:"version" binding:"required"`
	DeployedBy  string `json:"deployed_by"`
}

// Deployment status constants
const (
	DeploymentStatusPending   = "pending"
	DeploymentStatusDeploying = "deploying"
	DeploymentStatusSuccess   = "success"
	DeploymentStatusFailed    = "failed"
)

// Deployment environment constants
const (
	DeploymentEnvDev        = "dev"
	DeploymentEnvStaging    = "staging"
	DeploymentEnvProduction = "production"
)
