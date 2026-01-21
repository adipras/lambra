package service

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
)

// SanitizeServiceName applies naming rules to a service name:
// 1. Only alphanumeric characters and spaces allowed
// 2. Spaces are replaced with dashes
// 3. "svc-" prefix is added
// 4. Converts to lowercase
func SanitizeServiceName(name string) string {
	// Remove all characters except alphanumeric and spaces
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
	name = reg.ReplaceAllString(name, "")

	// Trim whitespace
	name = strings.TrimSpace(name)

	// Replace spaces with dashes
	name = strings.ReplaceAll(name, " ", "-")

	// Remove consecutive dashes
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}

	// Convert to lowercase
	name = strings.ToLower(name)

	// Add svc- prefix if not already present
	if !strings.HasPrefix(name, "svc-") {
		name = "svc-" + name
	}

	return name
}

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

// ValidateDBConnection tests if we can connect to the specified database
func (s *ProjectService) ValidateDBConnection(host string, port int, user, password, dbName string) error {
	// Build DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=5s", user, password, host, port)

	// Try to connect
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()

	// Set connection timeout
	db.SetConnMaxLifetime(5 * time.Second)

	// Ping to verify connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Try to create the database if it doesn't exist
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	return nil
}

func (s *ProjectService) CreateProject(req *models.CreateProjectRequest) (*models.Project, error) {
	// Validate database connection first
	if err := s.ValidateDBConnection(req.DBHost, req.DBPort, req.DBUser, req.DBPassword, req.DBName); err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	// Sanitize the service name (apply naming rules)
	sanitizedName := SanitizeServiceName(req.Name)

	project := &models.Project{
		Name:       sanitizedName,
		Namespace:  req.Namespace,
		Status:     models.ProjectStatusActive,
		DBHost:     req.DBHost,
		DBPort:     req.DBPort,
		DBUser:     req.DBUser,
		DBPassword: req.DBPassword,
		DBName:     req.DBName,
	}

	if req.Description != "" {
		project.Description = sql.NullString{String: req.Description, Valid: true}
	}

	// Set created_by (in future, get from auth context)
	project.SetCreatedBy("system")

	err := s.repo.Create(project)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

func (s *ProjectService) GetProjectByUUID(uuid string) (*models.Project, error) {
	return s.repo.GetByUUID(uuid)
}

func (s *ProjectService) GetProjectWithRelations(uuid string) (*models.ProjectWithRelations, error) {
	return s.repo.GetWithGitRepoByUUID(uuid)
}

func (s *ProjectService) GetAllProjects(page, limit int) ([]models.Project, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit
	return s.repo.GetAll(limit, offset)
}

func (s *ProjectService) UpdateProject(uuid string, req *models.UpdateProjectRequest) (*models.Project, error) {
	project, err := s.repo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.Status != "" {
		project.Status = req.Status
	}

	// Set updated_by (in future, get from auth context)
	project.SetUpdatedBy("system")

	err = s.repo.Update(project)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

func (s *ProjectService) DeleteProject(uuid string) error {
	_, err := s.repo.GetByUUID(uuid)
	if err != nil {
		return err
	}

	// Soft delete with deleted_by (in future, get from auth context)
	return s.repo.DeleteByUUID(uuid, "system")
}
