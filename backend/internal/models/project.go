package models

import (
	"database/sql"
	"encoding/json"
)

// Project represents a microservice project
type Project struct {
	BaseEntity
	Name        string         `db:"name" json:"name"`
	Description sql.NullString `db:"description" json:"-"`
	Status      string         `db:"status" json:"status"` // active, generating, failed, archived
	GitRepoID   sql.NullString `db:"git_repo_id" json:"-"` // UUID
	Namespace   string         `db:"namespace" json:"namespace"` // k8s namespace
}

// MarshalJSON custom JSON marshaling for Project
func (p Project) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		BaseEntityJSON
		Name        string  `json:"name"`
		Description string  `json:"description,omitempty"`
		Status      string  `json:"status"`
		GitRepoID   *string `json:"git_repo_id,omitempty"`
		Namespace   string  `json:"namespace"`
	}{
		BaseEntityJSON: p.BaseEntity.ToJSON(),
		Name:           p.Name,
		Description:    p.Description.String,
		Status:         p.Status,
		GitRepoID:      func() *string { if p.GitRepoID.Valid { return &p.GitRepoID.String }; return nil }(),
		Namespace:      p.Namespace,
	})
}

// ProjectWithRelations includes related data
type ProjectWithRelations struct {
	Project
	GitRepo     *GitRepository   `json:"git_repo,omitempty"`
	Entities    []Entity         `json:"entities,omitempty"`
	Deployments []Deployment     `json:"deployments,omitempty"`
}

// CreateProjectRequest for creating a new project
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
	Namespace   string `json:"namespace" binding:"required,min=3,max=50"`
}

// UpdateProjectRequest for updating project
type UpdateProjectRequest struct {
	Name        string `json:"name" binding:"omitempty,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
	Status      string `json:"status" binding:"omitempty,oneof=active generating failed archived"`
}

// ProjectStatus constants
const (
	ProjectStatusActive     = "active"
	ProjectStatusGenerating = "generating"
	ProjectStatusFailed     = "failed"
	ProjectStatusArchived   = "archived"
)
