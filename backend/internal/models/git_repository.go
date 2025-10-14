package models

import (
	"database/sql"
	"encoding/json"
)

// GitRepository represents a Git repository
type GitRepository struct {
	BaseEntity
	ProjectID        string         `db:"project_id" json:"project_id"`
	RepoURL          string         `db:"repo_url" json:"repo_url"`
	RepoName         string         `db:"repo_name" json:"repo_name"`
	GitLabRepoID     int64          `db:"gitlab_repo_id" json:"gitlab_repo_id"`
	DefaultBranch    string         `db:"default_branch" json:"default_branch"`
	DevelopBranch    string         `db:"develop_branch" json:"develop_branch"`
	StagingBranch    string         `db:"staging_branch" json:"staging_branch"`
	ProductionBranch string         `db:"production_branch" json:"production_branch"`
	LastCommitHash   sql.NullString `db:"last_commit_hash" json:"-"`
}

// MarshalJSON custom JSON marshaling for GitRepository
func (g GitRepository) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		BaseEntityJSON
		ProjectID        string `json:"project_id"`
		RepoURL          string `json:"repo_url"`
		RepoName         string `json:"repo_name"`
		GitLabRepoID     int64  `json:"gitlab_repo_id"`
		DefaultBranch    string `json:"default_branch"`
		DevelopBranch    string `json:"develop_branch"`
		StagingBranch    string `json:"staging_branch"`
		ProductionBranch string `json:"production_branch"`
		LastCommitHash   string `json:"last_commit_hash,omitempty"`
	}{
		BaseEntityJSON:   g.BaseEntity.ToJSON(),
		ProjectID:        g.ProjectID,
		RepoURL:          g.RepoURL,
		RepoName:         g.RepoName,
		GitLabRepoID:     g.GitLabRepoID,
		DefaultBranch:    g.DefaultBranch,
		DevelopBranch:    g.DevelopBranch,
		StagingBranch:    g.StagingBranch,
		ProductionBranch: g.ProductionBranch,
		LastCommitHash:   g.LastCommitHash.String,
	})
}

// CreateGitRepositoryRequest for creating Git repo
type CreateGitRepositoryRequest struct {
	ProjectID   string `json:"project_id" binding:"required"`
	RepoName    string `json:"repo_name" binding:"required,min=3,max=100"`
	Description string `json:"description"`
}

// GitBranches helper struct
type GitBranches struct {
	Develop    string
	Staging    string
	Production string
}
