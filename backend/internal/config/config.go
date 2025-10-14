package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	GitLab     GitLabConfig
	RBAC       RBACConfig
	Ambassador AmbassadorConfig
	Workspace  WorkspaceConfig
}

type ServerConfig struct {
	Port    string
	Env     string
	GinMode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type GitLabConfig struct {
	URL     string
	Token   string
	GroupID string
}

type RBACConfig struct {
	ServiceURL string
}

type AmbassadorConfig struct {
	ServiceURL string
}

type WorkspaceConfig struct {
	Path string
}

func Load() (*Config, error) {
	// Load .env file if exists (ignore error in production)
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			Env:     getEnv("ENV", "development"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "lambra_db"),
		},
		GitLab: GitLabConfig{
			URL:     getEnv("GITLAB_URL", "https://gitlab.com"),
			Token:   getEnv("GITLAB_TOKEN", ""),
			GroupID: getEnv("GITLAB_GROUP_ID", ""),
		},
		RBAC: RBACConfig{
			ServiceURL: getEnv("RBAC_SERVICE_URL", "http://localhost:8081"),
		},
		Ambassador: AmbassadorConfig{
			ServiceURL: getEnv("AMBASSADOR_SERVICE_URL", "http://localhost:8082"),
		},
		Workspace: WorkspaceConfig{
			Path: getEnv("WORKSPACE_PATH", "/tmp/lambra-workspace"),
		},
	}

	return config, nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Name)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
