package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Deployment represents a deployment instance
type Deployment struct {
	BaseEntity
	ProjectID     int64          `db:"project_id" json:"-"`
	ProjectUUID   string         `db:"-" json:"project_id"` // Virtual field for JSON
	SnapshotID    sql.NullInt64  `db:"snapshot_id" json:"-"`
	SnapshotUUID  string         `db:"-" json:"snapshot_id,omitempty"` // Virtual field for JSON
	Environment   string         `db:"environment" json:"environment"` // dev, staging, production
	Status        string         `db:"status" json:"status"`           // pending, deploying, success, failed
	Version       string         `db:"version" json:"version"`
	DeployedBy    sql.NullString `db:"deployed_by" json:"-"`
	DeploymentURL sql.NullString `db:"deployment_url" json:"-"`
	ErrorMessage  sql.NullString `db:"error_message" json:"-"`
	StartedAt     time.Time      `db:"started_at" json:"started_at"`
	CompletedAt   sql.NullTime   `db:"completed_at" json:"-"`
}

// MarshalJSON implements custom JSON marshaling for Deployment
func (d Deployment) MarshalJSON() ([]byte, error) {
	type Alias Deployment
	return json.Marshal(&struct {
		ID            string     `json:"id"`
		ProjectID     string     `json:"project_id"`
		SnapshotID    string     `json:"snapshot_id,omitempty"`
		DeployedBy    string     `json:"deployed_by,omitempty"`
		DeploymentURL string     `json:"deployment_url,omitempty"`
		ErrorMessage  string     `json:"error_message,omitempty"`
		CompletedAt   *time.Time `json:"completed_at,omitempty"`
		Alias
	}{
		ID:            d.UUID,
		ProjectID:     d.ProjectUUID,
		SnapshotID:    d.SnapshotUUID,
		DeployedBy:    nullStringToString(d.DeployedBy),
		DeploymentURL: nullStringToString(d.DeploymentURL),
		ErrorMessage:  nullStringToString(d.ErrorMessage),
		CompletedAt:   nullTimeToPtr(d.CompletedAt),
		Alias:         (Alias)(d),
	})
}

// Helper functions for null handling
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// DeploymentWithLogs includes deployment logs
type DeploymentWithLogs struct {
	Deployment
	Logs []DeploymentLog `json:"logs,omitempty"`
}

// DeploymentLog represents deployment log entries
type DeploymentLog struct {
	ID             int64     `db:"id" json:"id"`
	DeploymentID   int64     `db:"deployment_id" json:"-"`
	DeploymentUUID string    `db:"-" json:"deployment_id"` // Virtual field for JSON
	Level          string    `db:"level" json:"level"`     // info, warning, error, debug
	Message        string    `db:"message" json:"message"`
	Step           string    `db:"step" json:"step,omitempty"` // Optional step identifier
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// CreateDeploymentRequest for creating deployment
type CreateDeploymentRequest struct {
	ProjectUUID string `json:"project_id" binding:"required"`
	SnapshotID  int64  `json:"snapshot_id"`
	Environment string `json:"environment" binding:"required,oneof=dev staging production"`
	Version     string `json:"version"`
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

// Log level constants
const (
	LogLevelDebug   = "debug"
	LogLevelInfo    = "info"
	LogLevelWarning = "warning"
	LogLevelError   = "error"
)

// Deployment step constants
const (
	DeployStepInit         = "init"
	DeployStepSnapshot     = "snapshot"
	DeployStepGenerateCode = "generate_code"
	DeployStepWriteFiles   = "write_files"
	DeployStepDockerBuild  = "docker_build"
	DeployStepDockerStart  = "docker_start"
	DeployStepComplete     = "complete"
)
