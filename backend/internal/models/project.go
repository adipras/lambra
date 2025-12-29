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
	Status      string         `db:"status" json:"status"`       // active, generating, failed, archived
	GitRepoID   sql.NullInt64  `db:"git_repo_id" json:"-"`       // Foreign key to git_repositories.id
	Namespace   string         `db:"namespace" json:"namespace"` // k8s namespace
	// Database configuration for generated service
	DBHost     string `db:"db_host" json:"db_host"`
	DBPort     int    `db:"db_port" json:"db_port"`
	DBUser     string `db:"db_user" json:"db_user"`
	DBPassword string `db:"db_password" json:"-"` // Not exposed in JSON
	DBName     string `db:"db_name" json:"db_name"`
}

// MarshalJSON custom JSON marshaling for Project
func (p Project) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		BaseEntityJSON
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Status      string `json:"status"`
		Namespace   string `json:"namespace"`
		DBHost      string `json:"db_host"`
		DBPort      int    `json:"db_port"`
		DBUser      string `json:"db_user"`
		DBName      string `json:"db_name"`
	}{
		BaseEntityJSON: p.BaseEntity.ToJSON(),
		Name:           p.Name,
		Description:    p.Description.String,
		Status:         p.Status,
		Namespace:      p.Namespace,
		DBHost:         p.DBHost,
		DBPort:         p.DBPort,
		DBUser:         p.DBUser,
		DBName:         p.DBName,
	})
}

// ProjectWithRelations includes related data
type ProjectWithRelations struct {
	Project
	GitRepo     *GitRepository `json:"git_repo,omitempty"`
	Entities    []Entity       `json:"entities,omitempty"`
	Deployments []Deployment   `json:"deployments,omitempty"`
}

// CreateProjectRequest for creating a new project
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
	Namespace   string `json:"namespace" binding:"required,min=3,max=50"`
	// Database configuration (required for generated service)
	DBHost     string `json:"db_host" binding:"required"`
	DBPort     int    `json:"db_port" binding:"required,min=1,max=65535"`
	DBUser     string `json:"db_user" binding:"required"`
	DBPassword string `json:"db_password" binding:"required"`
	DBName     string `json:"db_name" binding:"required"`
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
