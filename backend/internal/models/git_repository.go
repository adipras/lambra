package models

import (
	"database/sql"
	"time"
)

// GitRepository represents a Git repository
type GitRepository struct {
	ID             int64          `db:"id" json:"id"`
	ProjectID      int64          `db:"project_id" json:"project_id"`
	RepoURL        string         `db:"repo_url" json:"repo_url"`
	RepoName       string         `db:"repo_name" json:"repo_name"`
	GitLabRepoID   int64          `db:"gitlab_repo_id" json:"gitlab_repo_id"`
	DefaultBranch  string         `db:"default_branch" json:"default_branch"`
	DevelopBranch  string         `db:"develop_branch" json:"develop_branch"`
	StagingBranch  string         `db:"staging_branch" json:"staging_branch"`
	ProductionBranch string       `db:"production_branch" json:"production_branch"`
	LastCommitHash sql.NullString `db:"last_commit_hash" json:"last_commit_hash,omitempty"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at" json:"updated_at"`
}

// CreateGitRepositoryRequest for creating Git repo
type CreateGitRepositoryRequest struct {
	ProjectID   int64  `json:"project_id" binding:"required"`
	RepoName    string `json:"repo_name" binding:"required,min=3,max=100"`
	Description string `json:"description"`
}

// GitBranches helper struct
type GitBranches struct {
	Develop    string
	Staging    string
	Production string
}
