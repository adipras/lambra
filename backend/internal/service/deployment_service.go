package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
)

// DeploymentService handles service deployment operations
type DeploymentService struct {
	projectRepo    *repository.ProjectRepository
	entityRepo     *repository.EntityRepository
	endpointRepo   *repository.EndpointRepository
	snapshotRepo   *repository.SnapshotRepository
	generatorSvc   *GeneratorService
	workspacePath  string
	basePort       int
}

// NewDeploymentService creates a new deployment service
func NewDeploymentService(
	projectRepo *repository.ProjectRepository,
	entityRepo *repository.EntityRepository,
	endpointRepo *repository.EndpointRepository,
	snapshotRepo *repository.SnapshotRepository,
	generatorSvc *GeneratorService,
	workspacePath string,
) *DeploymentService {
	return &DeploymentService{
		projectRepo:   projectRepo,
		entityRepo:    entityRepo,
		endpointRepo:  endpointRepo,
		snapshotRepo:  snapshotRepo,
		generatorSvc:  generatorSvc,
		workspacePath: workspacePath,
		basePort:      9000, // Generated services start from port 9000
	}
}

// ServiceStatus represents the status of a deployed service
type ServiceStatus struct {
	ProjectID   string `json:"project_id"`
	ServiceName string `json:"service_name"`
	Status      string `json:"status"` // running, stopped, not_deployed
	Port        int    `json:"port,omitempty"`
	URL         string `json:"url,omitempty"`          // External URL for frontend display (localhost)
	InternalURL string `json:"internal_url,omitempty"` // Internal URL for backend testing (host.docker.internal)
	StartedAt   string `json:"started_at,omitempty"`
	Error       string `json:"error,omitempty"`
}

// DeployResult represents the result of a deployment
type DeployResult struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ServiceName string `json:"service_name"`
	Port        int    `json:"port"`
	URL         string `json:"url"`
	InternalURL string `json:"internal_url"`
}

// DeployProject generates code and deploys the service
func (s *DeploymentService) DeployProject(ctx context.Context, projectUUID string) (*DeployResult, error) {
	// Get project
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Auto-create snapshot before deployment
	snapshot, err := s.createSnapshot(project)
	if err != nil {
		// Log warning but continue with deployment
		fmt.Printf("Warning: failed to create snapshot: %v\n", err)
	} else {
		fmt.Printf("Snapshot created: %s (version %s)\n", snapshot.UUID, snapshot.Version)
	}

	serviceName := toKebabCase(project.Name)
	serviceDir := filepath.Join(s.workspacePath, serviceName)

	// Create service directory
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create service directory: %w", err)
	}

	// Generate code (pass empty string so file paths are relative)
	response, err := s.generatorSvc.GenerateProjectByUUID(ctx, projectUUID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate code: %w", err)
	}

	// Write generated files
	for _, file := range response.Files {
		filePath := filepath.Join(serviceDir, file.Path)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory for %s: %w", file.Path, err)
		}
		if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}
	}

	// Assign port based on project ID
	port := s.getPortForProject(project.ID)

	// Generate Docker files
	if err := s.generateDockerFiles(project, serviceDir, port); err != nil {
		return nil, fmt.Errorf("failed to generate Docker files: %w", err)
	}

	// Generate main.go and go.mod
	if err := s.generateGoFiles(project, serviceDir, port); err != nil {
		return nil, fmt.Errorf("failed to generate Go files: %w", err)
	}

	// Build and start the service
	if err := s.startService(serviceDir); err != nil {
		return nil, fmt.Errorf("failed to start service: %w", err)
	}

	return &DeployResult{
		Success:     true,
		Message:     fmt.Sprintf("Service %s deployed successfully", serviceName),
		ServiceName: serviceName,
		Port:        port,
		URL:         fmt.Sprintf("http://localhost:%d", port),
		InternalURL: fmt.Sprintf("http://host.docker.internal:%d", port),
	}, nil
}

// StartService starts a deployed service
func (s *DeploymentService) StartService(ctx context.Context, projectUUID string) (*ServiceStatus, error) {
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	serviceName := toKebabCase(project.Name)
	serviceDir := filepath.Join(s.workspacePath, serviceName)

	// Check if service directory exists
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("service not deployed, please deploy first")
	}

	if err := s.startService(serviceDir); err != nil {
		return nil, fmt.Errorf("failed to start service: %w", err)
	}

	port := s.getPortForProject(project.ID)
	return &ServiceStatus{
		ProjectID:   projectUUID,
		ServiceName: serviceName,
		Status:      "running",
		Port:        port,
		URL:         fmt.Sprintf("http://localhost:%d", port),
		InternalURL: fmt.Sprintf("http://host.docker.internal:%d", port),
		StartedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

// StopService stops a running service
func (s *DeploymentService) StopService(ctx context.Context, projectUUID string) (*ServiceStatus, error) {
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	serviceName := toKebabCase(project.Name)
	serviceDir := filepath.Join(s.workspacePath, serviceName)

	// Check if service directory exists
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return &ServiceStatus{
			ProjectID:   projectUUID,
			ServiceName: serviceName,
			Status:      "not_deployed",
		}, nil
	}

	if err := s.stopService(serviceDir); err != nil {
		return nil, fmt.Errorf("failed to stop service: %w", err)
	}

	return &ServiceStatus{
		ProjectID:   projectUUID,
		ServiceName: serviceName,
		Status:      "stopped",
	}, nil
}

// RedeployService redeploys a service (regenerate code + down + up)
func (s *DeploymentService) RedeployService(ctx context.Context, projectUUID string) (*ServiceStatus, error) {
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	serviceName := toKebabCase(project.Name)
	serviceDir := filepath.Join(s.workspacePath, serviceName)

	// Check if service directory exists
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("service not deployed, please deploy first")
	}

	// Stop and remove containers (docker compose down)
	if err := s.destroyService(serviceDir); err != nil {
		return nil, fmt.Errorf("failed to stop service: %w", err)
	}

	// Regenerate code (pass empty string so file paths are relative)
	response, err := s.generatorSvc.GenerateProjectByUUID(ctx, projectUUID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to regenerate code: %w", err)
	}

	// Write regenerated files
	for _, file := range response.Files {
		filePath := filepath.Join(serviceDir, file.Path)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory for %s: %w", file.Path, err)
		}
		if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}
	}

	port := s.getPortForProject(project.ID)

	// Regenerate Docker files
	if err := s.generateDockerFiles(project, serviceDir, port); err != nil {
		return nil, fmt.Errorf("failed to regenerate Docker files: %w", err)
	}

	// Regenerate main.go with routes
	if err := s.generateGoFiles(project, serviceDir, port); err != nil {
		return nil, fmt.Errorf("failed to regenerate Go files: %w", err)
	}

	// Rebuild and start (docker compose up -d --build)
	if err := s.startService(serviceDir); err != nil {
		return nil, fmt.Errorf("failed to restart service: %w", err)
	}

	return &ServiceStatus{
		ProjectID:   projectUUID,
		ServiceName: serviceName,
		Status:      "running",
		Port:        port,
		URL:         fmt.Sprintf("http://localhost:%d", port),
		InternalURL: fmt.Sprintf("http://host.docker.internal:%d", port),
		StartedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

// GetServiceStatus gets the status of a service
func (s *DeploymentService) GetServiceStatus(ctx context.Context, projectUUID string) (*ServiceStatus, error) {
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	serviceName := toKebabCase(project.Name)
	serviceDir := filepath.Join(s.workspacePath, serviceName)

	// Check if service directory exists
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return &ServiceStatus{
			ProjectID:   projectUUID,
			ServiceName: serviceName,
			Status:      "not_deployed",
		}, nil
	}

	// Check if container is running
	status := s.checkContainerStatus(serviceName)
	port := s.getPortForProject(project.ID)

	result := &ServiceStatus{
		ProjectID:   projectUUID,
		ServiceName: serviceName,
		Status:      status,
	}

	if status == "running" {
		result.Port = port
		result.URL = fmt.Sprintf("http://localhost:%d", port)
		result.InternalURL = fmt.Sprintf("http://host.docker.internal:%d", port)
	}

	return result, nil
}

// DestroyServiceCompletely stops containers, removes volumes, deletes workspace, and removes from database
func (s *DeploymentService) DestroyServiceCompletely(ctx context.Context, projectUUID string) error {
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	serviceName := toKebabCase(project.Name)
	serviceDir := filepath.Join(s.workspacePath, serviceName)

	// 1. Stop and remove Docker containers with volumes (if service directory exists)
	if _, err := os.Stat(serviceDir); err == nil {
		if err := s.destroyService(serviceDir); err != nil {
			// Log but continue - containers might not exist
			fmt.Printf("Warning: failed to destroy containers: %v\n", err)
		}

		// 2. Remove workspace directory completely
		if err := os.RemoveAll(serviceDir); err != nil {
			fmt.Printf("Warning: failed to remove workspace directory: %v\n", err)
		}
	}

	// 3. Delete all endpoints for this project from database
	if err := s.endpointRepo.HardDeleteByProjectID(project.ID); err != nil {
		fmt.Printf("Warning: failed to delete endpoints: %v\n", err)
	}

	// 4. Delete all entities for this project from database
	if err := s.entityRepo.HardDeleteByProjectID(project.ID); err != nil {
		fmt.Printf("Warning: failed to delete entities: %v\n", err)
	}

	// 5. Delete the project from database
	if err := s.projectRepo.HardDeleteByUUID(projectUUID); err != nil {
		return fmt.Errorf("failed to delete project from database: %w", err)
	}

	return nil
}

// createSnapshot creates a snapshot of the current project state before deployment
func (s *DeploymentService) createSnapshot(project *models.Project) (*models.GenerationSnapshot, error) {
	// Get all entities for the project
	entities, err := s.entityRepo.GetByProjectID(project.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	// Get all endpoints for each entity
	var allEndpoints []models.Endpoint
	for _, entity := range entities {
		endpoints, err := s.endpointRepo.GetByEntityID(entity.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get endpoints for entity %s: %w", entity.Name, err)
		}
		allEndpoints = append(allEndpoints, endpoints...)
	}

	// Create snapshot metadata
	metadata := models.SnapshotMetadata{
		Entities:  entities,
		Endpoints: allEndpoints,
		Config: map[string]interface{}{
			"namespace": project.Namespace,
			"db_host":   project.DBHost,
			"db_port":   project.DBPort,
			"db_user":   project.DBUser,
			"db_name":   project.DBName,
		},
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create database snapshot info
	dbSnapshot := models.DatabaseSnapshotInfo{
		MigrationVersion:  fmt.Sprintf("v%s", time.Now().Format("20060102150405")),
		AppliedMigrations: []string{},
	}

	dbSnapshotJSON, err := json.Marshal(dbSnapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal database snapshot: %w", err)
	}

	// Count existing snapshots for version numbering
	count, err := s.snapshotRepo.CountByProjectID(project.ID)
	if err != nil {
		count = 0
	}

	// Generate version
	version := fmt.Sprintf("v1.%d.0", count+1)

	// Deactivate previous active snapshot
	if err := s.snapshotRepo.SetAllInactiveByProjectID(project.ID, "deployment"); err != nil {
		fmt.Printf("Warning: failed to deactivate old snapshots: %v\n", err)
	}

	// Create snapshot
	snapshot := &models.GenerationSnapshot{
		ProjectID:        project.ID,
		Version:          version,
		GitCommitHash:    fmt.Sprintf("deploy-%d", time.Now().Unix()),
		Metadata:         metadataJSON,
		DatabaseSnapshot: dbSnapshotJSON,
		Status:           models.SnapshotStatusActive,
	}
	snapshot.CreatedBy.String = "deployment"
	snapshot.CreatedBy.Valid = true

	if err := s.snapshotRepo.Create(snapshot); err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return snapshot, nil
}

// Helper methods

func (s *DeploymentService) getPortForProject(projectID int64) int {
	// Simple port assignment based on project ID
	return s.basePort + int(projectID%1000)
}

func (s *DeploymentService) startService(serviceDir string) error {
	cmd := exec.Command("docker", "compose", "up", "-d", "--build")
	cmd.Dir = serviceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *DeploymentService) stopService(serviceDir string) error {
	cmd := exec.Command("docker", "compose", "stop")
	cmd.Dir = serviceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *DeploymentService) destroyService(serviceDir string) error {
	cmd := exec.Command("docker", "compose", "down", "-v")
	cmd.Dir = serviceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *DeploymentService) checkContainerStatus(serviceName string) string {
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", serviceName), "--format", "{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		return "stopped"
	}

	outputStr := strings.TrimSpace(string(output))
	if strings.Contains(outputStr, "Up") {
		return "running"
	}
	return "stopped"
}

func (s *DeploymentService) generateDockerFiles(project *models.Project, serviceDir string, port int) error {
	serviceName := toKebabCase(project.Name)

	// Use project's database configuration (no MySQL container needed)
	data := map[string]any{
		"ServiceName":      serviceName,
		"Port":             port,
		"DBHost":           project.DBHost,
		"DBPort":           project.DBPort,
		"DBUser":           project.DBUser,
		"DBPassword":       project.DBPassword,
		"DBName":           project.DBName,
		"Environment":      "development",
		"GinMode":          "debug",
	}

	// Generate docker-compose.yml (no MySQL container, connects to external DB)
	dockerComposeTmpl := `services:
  {{.ServiceName}}:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: {{.ServiceName}}
    ports:
      - "{{.Port}}:{{.Port}}"
    environment:
      PORT: {{.Port}}
      ENV: {{.Environment}}
      GIN_MODE: {{.GinMode}}
      DB_HOST: {{.DBHost}}
      DB_PORT: {{.DBPort}}
      DB_USER: {{.DBUser}}
      DB_PASSWORD: {{.DBPassword}}
      DB_NAME: {{.DBName}}
    extra_hosts:
      - "host.docker.internal:host-gateway"
    networks:
      - {{.ServiceName}}-network
    restart: unless-stopped

networks:
  {{.ServiceName}}-network:
    driver: bridge
`

	dockerfileTmpl := `FROM golang:1.21-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod ./
COPY go.sum* ./
RUN go mod download || true
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE {{.Port}}
CMD ["./main"]
`

	// Write docker-compose.yml
	if err := s.renderTemplate(dockerComposeTmpl, data, filepath.Join(serviceDir, "docker-compose.yml")); err != nil {
		return err
	}

	// Write Dockerfile
	if err := s.renderTemplate(dockerfileTmpl, data, filepath.Join(serviceDir, "Dockerfile")); err != nil {
		return err
	}

	return nil
}

func (s *DeploymentService) generateGoFiles(project *models.Project, serviceDir string, port int) error {
	serviceName := toKebabCase(project.Name)

	// Get entities for this project
	entities, err := s.entityRepo.GetByProjectID(project.ID)
	if err != nil {
		return fmt.Errorf("failed to get entities: %w", err)
	}

	// go.mod
	goModContent := fmt.Sprintf(`module %s

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/go-sql-driver/mysql v1.7.1
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.3.5
)
`, serviceName)

	if err := os.WriteFile(filepath.Join(serviceDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		return err
	}

	// Create cmd/server directory
	cmdDir := filepath.Join(serviceDir, "cmd", "server")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return err
	}

	// Build imports, initializations, routes, and migrations for each entity
	var imports strings.Builder
	var inits strings.Builder
	var routes strings.Builder
	var migrations strings.Builder

	imports.WriteString(fmt.Sprintf("\t\"%s/repository\"\n", serviceName))
	imports.WriteString(fmt.Sprintf("\t\"%s/service\"\n", serviceName))
	imports.WriteString(fmt.Sprintf("\t\"%s/api/handlers\"\n", serviceName))

	for _, entity := range entities {
		// Ensure entity name is PascalCase for type names
		entityNamePascal := toPascalCase(entity.Name)
		entityNameLower := strings.ToLower(entityNamePascal[:1]) + entityNamePascal[1:]
		entityNameSnake := toSnakeCase(entity.Name)

		// Repository, Service, Handler initialization
		inits.WriteString(fmt.Sprintf("\t%sRepo := repository.New%sRepository(db)\n", entityNameLower, entityNamePascal))
		inits.WriteString(fmt.Sprintf("\t%sSvc := service.New%sService(%sRepo)\n", entityNameLower, entityNamePascal, entityNameLower))
		inits.WriteString(fmt.Sprintf("\t%sHandler := handlers.New%sHandler(%sSvc)\n\n", entityNameLower, entityNamePascal, entityNameLower))

		// Generate migration SQL for this entity
		migrationSQL := s.generateMigrationSQL(entity)
		migrations.WriteString(migrationSQL)

		// Get endpoints for this entity
		endpoints, err := s.endpointRepo.GetByEntityID(entity.ID)
		if err != nil {
			continue
		}

		// Generate routes for each endpoint
		for _, endpoint := range endpoints {
			// Convert endpoint name to valid Go method name (PascalCase, no spaces)
			handlerMethod := toPascalCase(endpoint.Name)
			routes.WriteString(fmt.Sprintf("\tr.%s(\"%s\", %sHandler.%s)\n",
				endpoint.Method, endpoint.Path, entityNameLower, handlerMethod))
		}

		// Add default CRUD routes if no endpoints exist
		if len(endpoints) == 0 {
			basePath := "/" + entityNameSnake + "s"
			entityNamePlural := pluralize(entityNamePascal)
			routes.WriteString(fmt.Sprintf("\tr.GET(\"%s\", %sHandler.List%s)\n", basePath, entityNameLower, entityNamePlural))
			routes.WriteString(fmt.Sprintf("\tr.GET(\"%s/:id\", %sHandler.Get%s)\n", basePath, entityNameLower, entityNamePascal))
			routes.WriteString(fmt.Sprintf("\tr.POST(\"%s\", %sHandler.Create%s)\n", basePath, entityNameLower, entityNamePascal))
			routes.WriteString(fmt.Sprintf("\tr.PUT(\"%s/:id\", %sHandler.Update%s)\n", basePath, entityNameLower, entityNamePascal))
			routes.WriteString(fmt.Sprintf("\tr.DELETE(\"%s/:id\", %sHandler.Delete%s)\n", basePath, entityNameLower, entityNamePascal))
		}
	}

	// main.go with auto-migration
	mainGoContent := fmt.Sprintf(`package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
%s)

// runMigrations creates tables if they don't exist
func runMigrations(db *sqlx.DB) error {
	migrations := []string{
%s	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %%w", err)
		}
	}
	return nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "%d"
	}

	// Database connection
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%%s:%%s@tcp(%%s:%%s)/%%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %%v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	// Run auto-migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %%v", err)
	}
	log.Println("Migrations completed successfully")

	// Initialize repositories, services, and handlers
%s
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "%s",
		})
	})

	// Entity routes
%s
	log.Printf("Server starting on port %%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
`, imports.String(), migrations.String(), port, inits.String(), serviceName, routes.String())

	if err := os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(mainGoContent), 0644); err != nil {
		return err
	}

	// Run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = serviceDir
	cmd.Run() // Ignore error, it might fail without network

	return nil
}

// generateMigrationSQL generates CREATE TABLE statement for an entity
func (s *DeploymentService) generateMigrationSQL(entity models.Entity) string {
	// Parse fields from JSON
	var fields []models.EntityField
	if err := json.Unmarshal(entity.Fields, &fields); err != nil {
		return ""
	}

	var sql strings.Builder
	tableName := entity.TableName

	sql.WriteString(fmt.Sprintf("\t\t`CREATE TABLE IF NOT EXISTS %s (\n", tableName))
	sql.WriteString("\t\t\tid BIGINT PRIMARY KEY AUTO_INCREMENT,\n")
	sql.WriteString("\t\t\tuuid CHAR(36) NOT NULL UNIQUE,\n")

	for _, field := range fields {
		columnName := toSnakeCase(field.Name)
		columnType := s.getSQLType(field.Type, field.Length)
		notNull := ""
		if field.Required {
			notNull = " NOT NULL"
		}
		defaultVal := ""
		if field.DefaultValue != "" {
			defaultVal = fmt.Sprintf(" DEFAULT %s", field.DefaultValue)
		}
		sql.WriteString(fmt.Sprintf("\t\t\t%s %s%s%s,\n", columnName, columnType, notNull, defaultVal))
	}

	sql.WriteString("\t\t\tcreated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,\n")
	sql.WriteString("\t\t\tupdated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n")
	sql.WriteString("\t\t\tdeleted_at TIMESTAMP NULL DEFAULT NULL,\n")
	sql.WriteString(fmt.Sprintf("\t\t\tINDEX idx_%s_uuid (uuid),\n", tableName))
	sql.WriteString(fmt.Sprintf("\t\t\tINDEX idx_%s_deleted_at (deleted_at)\n", tableName))
	sql.WriteString("\t\t)`,\n")

	return sql.String()
}

// getSQLType converts field type to MySQL type
func (s *DeploymentService) getSQLType(fieldType string, length int) string {
	switch fieldType {
	case "string":
		if length > 0 {
			return fmt.Sprintf("VARCHAR(%d)", length)
		}
		return "VARCHAR(255)"
	case "text":
		return "TEXT"
	case "int", "integer":
		return "INT"
	case "int64", "bigint":
		return "BIGINT"
	case "float", "float32":
		return "FLOAT"
	case "float64", "double":
		return "DOUBLE"
	case "bool", "boolean":
		return "BOOLEAN"
	case "date":
		return "DATE"
	case "datetime", "timestamp":
		return "DATETIME"
	case "json":
		return "JSON"
	default:
		return "VARCHAR(255)"
	}
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, r+32) // Convert to lowercase
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func toPascalCase(s string) string {
	if s == "" {
		return s
	}
	// If no delimiters (space, underscore, dash), just capitalize first letter
	if !strings.ContainsAny(s, " _-") {
		return strings.ToUpper(s[:1]) + s[1:]
	}
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '_' || r == '-'
	})
	for i := range words {
		if len(words[i]) > 0 {
			words[i] = strings.ToUpper(words[i][:1]) + strings.ToLower(words[i][1:])
		}
	}
	return strings.Join(words, "")
}

func (s *DeploymentService) renderTemplate(tmplStr string, data interface{}, outputPath string) error {
	tmpl, err := template.New("template").Parse(tmplStr)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

func toKebabCase(s string) string {
	var result []rune
	var lastWasDash bool
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 && !lastWasDash {
				result = append(result, '-')
			}
			result = append(result, r+32) // Convert to lowercase
			lastWasDash = false
		} else if r == ' ' || r == '_' {
			if !lastWasDash {
				result = append(result, '-')
				lastWasDash = true
			}
		} else {
			result = append(result, r)
			lastWasDash = false
		}
	}
	return string(result)
}

// pluralize converts a word to plural form
func pluralize(s string) string {
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	return s + "s"
}
