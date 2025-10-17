package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// GenerationSnapshot represents a snapshot of generated code
type GenerationSnapshot struct {
	ID               int64           `db:"id" json:"id"`
	ProjectID        int64           `db:"project_id" json:"project_id"`
	Version          string          `db:"version" json:"version"`
	GitCommitHash    string          `db:"git_commit_hash" json:"git_commit_hash"`
	GitTag           sql.NullString  `db:"git_tag" json:"git_tag,omitempty"`
	Metadata         json.RawMessage `db:"metadata" json:"metadata"`                   // Stores entities, endpoints config
	DatabaseSnapshot json.RawMessage `db:"database_snapshot" json:"database_snapshot"` // Migration info
	Status           string          `db:"status" json:"status"`                       // created, active, rolled_back
	CreatedBy        sql.NullString  `db:"created_by" json:"created_by,omitempty"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
}

// SnapshotMetadata contains snapshot metadata
type SnapshotMetadata struct {
	Entities  []Entity               `json:"entities"`
	Endpoints []Endpoint             `json:"endpoints"`
	Config    map[string]interface{} `json:"config"`
}

// DatabaseSnapshotInfo contains database migration info
type DatabaseSnapshotInfo struct {
	MigrationVersion  string   `json:"migration_version"`
	AppliedMigrations []string `json:"applied_migrations"`
}

// CreateSnapshotRequest for creating snapshot
type CreateSnapshotRequest struct {
	ProjectID        int64                `json:"project_id" binding:"required"`
	Version          string               `json:"version" binding:"required"`
	GitCommitHash    string               `json:"git_commit_hash" binding:"required"`
	GitTag           string               `json:"git_tag"`
	Metadata         SnapshotMetadata     `json:"metadata" binding:"required"`
	DatabaseSnapshot DatabaseSnapshotInfo `json:"database_snapshot"`
	CreatedBy        string               `json:"created_by"`
}

// Snapshot status constants
const (
	SnapshotStatusCreated    = "created"
	SnapshotStatusActive     = "active"
	SnapshotStatusRolledBack = "rolled_back"
)
