package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
)

// DeploymentService handles service deployment operations
type DeploymentService struct {
	projectRepo       *repository.ProjectRepository
	entityRepo        *repository.EntityRepository
	endpointRepo      *repository.EndpointRepository
	relationRepo      *repository.RelationRepository
	snapshotRepo      *repository.SnapshotRepository
	deploymentRepo    *repository.DeploymentRepository
	deploymentLogRepo *repository.DeploymentLogRepository
	generatorSvc      *GeneratorService
	workspacePath     string
	basePort          int
}

// NewDeploymentService creates a new deployment service
func NewDeploymentService(
	projectRepo *repository.ProjectRepository,
	entityRepo *repository.EntityRepository,
	endpointRepo *repository.EndpointRepository,
	relationRepo *repository.RelationRepository,
	snapshotRepo *repository.SnapshotRepository,
	deploymentRepo *repository.DeploymentRepository,
	deploymentLogRepo *repository.DeploymentLogRepository,
	generatorSvc *GeneratorService,
	workspacePath string,
) *DeploymentService {
	return &DeploymentService{
		projectRepo:       projectRepo,
		entityRepo:        entityRepo,
		endpointRepo:      endpointRepo,
		relationRepo:      relationRepo,
		snapshotRepo:      snapshotRepo,
		deploymentRepo:    deploymentRepo,
		deploymentLogRepo: deploymentLogRepo,
		generatorSvc:      generatorSvc,
		workspacePath:     workspacePath,
		basePort:          9000, // Generated services start from port 9000
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
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	DeploymentID string `json:"deployment_id"`
	ServiceName  string `json:"service_name"`
	Port         int    `json:"port"`
	URL          string `json:"url"`
	InternalURL  string `json:"internal_url"`
}

// DeployOptions contains options for deployment
type DeployOptions struct {
	ResetDatabase bool `json:"reset_database"` // If true, drop all tables before creating
}

// DeployProject generates code and deploys the service
func (s *DeploymentService) DeployProject(ctx context.Context, projectUUID string, opts *DeployOptions) (*DeployResult, error) {
	if opts == nil {
		opts = &DeployOptions{}
	}
	// Get project
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	serviceName := toKebabCase(project.Name)
	port := s.getPortForProject(project.ID)

	// Create deployment record
	deployment := &models.Deployment{
		ProjectID:   project.ID,
		Environment: models.DeploymentEnvDev,
		Status:      models.DeploymentStatusDeploying,
		Version:     fmt.Sprintf("v%d", time.Now().Unix()),
		StartedAt:   time.Now(),
	}
	deployment.CreatedBy = sql.NullString{String: "system", Valid: true}

	if err := s.deploymentRepo.Create(deployment); err != nil {
		return nil, fmt.Errorf("failed to create deployment record: %w", err)
	}

	// Helper function to log deployment steps
	logStep := func(level, message, step string) {
		s.deploymentLogRepo.Create(&models.DeploymentLog{
			DeploymentID: deployment.ID,
			Level:        level,
			Message:      message,
			Step:         step,
		})
	}

	logStep(models.LogLevelInfo, fmt.Sprintf("Starting deployment for project: %s", project.Name), models.DeployStepInit)

	// Get the previous snapshot to detect deleted entities
	var tablesToDrop []string
	if !opts.ResetDatabase {
		previousSnapshot, _ := s.snapshotRepo.GetLatestByProjectID(project.ID)
		if previousSnapshot != nil {
			tablesToDrop = s.detectDeletedTables(previousSnapshot, project.ID)
			if len(tablesToDrop) > 0 {
				logStep(models.LogLevelInfo, fmt.Sprintf("Detected %d deleted tables: %v", len(tablesToDrop), tablesToDrop), models.DeployStepInit)
			}
		}
	} else {
		logStep(models.LogLevelInfo, "Reset database option enabled - all tables will be dropped", models.DeployStepInit)
	}

	// Auto-create snapshot before deployment
	logStep(models.LogLevelInfo, "Creating snapshot...", models.DeployStepSnapshot)
	snapshot, err := s.createSnapshot(project)
	if err != nil {
		logStep(models.LogLevelWarning, fmt.Sprintf("Failed to create snapshot: %v", err), models.DeployStepSnapshot)
	} else {
		logStep(models.LogLevelInfo, fmt.Sprintf("Snapshot created: %s (version %s)", snapshot.UUID, snapshot.Version), models.DeployStepSnapshot)
		// Update deployment with snapshot ID
		deployment.SnapshotID = sql.NullInt64{Int64: snapshot.ID, Valid: true}
	}

	// Prepare migration options
	migOpts := &MigrationOptions{
		ResetDatabase: opts.ResetDatabase,
		TablesToDrop:  tablesToDrop,
	}

	serviceDir := filepath.Join(s.workspacePath, serviceName)

	// Create service directory
	logStep(models.LogLevelInfo, fmt.Sprintf("Creating service directory: %s", serviceDir), models.DeployStepGenerateCode)
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		logStep(models.LogLevelError, fmt.Sprintf("Failed to create service directory: %v", err), models.DeployStepGenerateCode)
		s.deploymentRepo.SetFailed(deployment.UUID, err.Error())
		return nil, fmt.Errorf("failed to create service directory: %w", err)
	}

	// Generate code (pass empty string so file paths are relative)
	logStep(models.LogLevelInfo, "Generating code...", models.DeployStepGenerateCode)
	response, err := s.generatorSvc.GenerateProjectByUUID(ctx, projectUUID, "")
	if err != nil {
		logStep(models.LogLevelError, fmt.Sprintf("Failed to generate code: %v", err), models.DeployStepGenerateCode)
		s.deploymentRepo.SetFailed(deployment.UUID, err.Error())
		return nil, fmt.Errorf("failed to generate code: %w", err)
	}
	logStep(models.LogLevelInfo, fmt.Sprintf("Generated %d files", len(response.Files)), models.DeployStepGenerateCode)

	// Write generated files
	logStep(models.LogLevelInfo, "Writing generated files...", models.DeployStepWriteFiles)
	for _, file := range response.Files {
		filePath := filepath.Join(serviceDir, file.Path)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			logStep(models.LogLevelError, fmt.Sprintf("Failed to create directory for %s: %v", file.Path, err), models.DeployStepWriteFiles)
			s.deploymentRepo.SetFailed(deployment.UUID, err.Error())
			return nil, fmt.Errorf("failed to create directory for %s: %w", file.Path, err)
		}
		if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
			logStep(models.LogLevelError, fmt.Sprintf("Failed to write file %s: %v", file.Path, err), models.DeployStepWriteFiles)
			s.deploymentRepo.SetFailed(deployment.UUID, err.Error())
			return nil, fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}
	}
	logStep(models.LogLevelInfo, fmt.Sprintf("Wrote %d files to %s", len(response.Files), serviceDir), models.DeployStepWriteFiles)

	// Generate Docker files
	logStep(models.LogLevelInfo, "Generating Docker files...", models.DeployStepDockerBuild)
	if err := s.generateDockerFiles(project, serviceDir, port); err != nil {
		logStep(models.LogLevelError, fmt.Sprintf("Failed to generate Docker files: %v", err), models.DeployStepDockerBuild)
		s.deploymentRepo.SetFailed(deployment.UUID, err.Error())
		return nil, fmt.Errorf("failed to generate Docker files: %w", err)
	}

	// Generate main.go and go.mod
	logStep(models.LogLevelInfo, "Generating Go files...", models.DeployStepDockerBuild)
	if err := s.generateGoFiles(project, serviceDir, port, migOpts); err != nil {
		logStep(models.LogLevelError, fmt.Sprintf("Failed to generate Go files: %v", err), models.DeployStepDockerBuild)
		s.deploymentRepo.SetFailed(deployment.UUID, err.Error())
		return nil, fmt.Errorf("failed to generate Go files: %w", err)
	}

	// Build and start the service
	logStep(models.LogLevelInfo, "Building and starting Docker container...", models.DeployStepDockerStart)
	if err := s.startService(serviceDir); err != nil {
		logStep(models.LogLevelError, fmt.Sprintf("Failed to start service: %v", err), models.DeployStepDockerStart)
		s.deploymentRepo.SetFailed(deployment.UUID, err.Error())
		return nil, fmt.Errorf("failed to start service: %w", err)
	}

	// Mark deployment as successful
	deploymentURL := fmt.Sprintf("http://localhost:%d", port)
	s.deploymentRepo.SetCompleted(deployment.UUID, models.DeploymentStatusSuccess, deploymentURL)
	logStep(models.LogLevelInfo, fmt.Sprintf("Service deployed successfully at %s", deploymentURL), models.DeployStepComplete)

	return &DeployResult{
		Success:      true,
		Message:      fmt.Sprintf("Service %s deployed successfully", serviceName),
		DeploymentID: deployment.UUID,
		ServiceName:  serviceName,
		Port:         port,
		URL:          deploymentURL,
		InternalURL:  fmt.Sprintf("http://host.docker.internal:%d", port),
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

	// Detect deleted tables by comparing with latest snapshot
	var tablesToDrop []string
	previousSnapshot, _ := s.snapshotRepo.GetLatestByProjectID(project.ID)
	if previousSnapshot != nil {
		tablesToDrop = s.detectDeletedTables(previousSnapshot, project.ID)
		if len(tablesToDrop) > 0 {
			fmt.Printf("Redeploy: detected %d deleted tables: %v\n", len(tablesToDrop), tablesToDrop)
		}
	}

	// Prepare migration options
	migOpts := &MigrationOptions{
		TablesToDrop: tablesToDrop,
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

	// Regenerate main.go with routes and proper migration options
	if err := s.generateGoFiles(project, serviceDir, port, migOpts); err != nil {
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

// DestroyServiceCompletely stops containers, removes volumes, deletes workspace, drops database, and removes from database
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

	// 3. Drop the generated service's database
	if err := s.dropServiceDatabase(project); err != nil {
		fmt.Printf("Warning: failed to drop service database: %v\n", err)
	}

	// 4. Delete all snapshots for this project from database
	if err := s.snapshotRepo.HardDeleteByProjectID(project.ID); err != nil {
		fmt.Printf("Warning: failed to delete snapshots: %v\n", err)
	}

	// 5. Delete all deployments for this project from database
	if err := s.deploymentRepo.HardDeleteByProjectID(project.ID); err != nil {
		fmt.Printf("Warning: failed to delete deployments: %v\n", err)
	}

	// 6. Delete all endpoints for this project from database
	if err := s.endpointRepo.HardDeleteByProjectID(project.ID); err != nil {
		fmt.Printf("Warning: failed to delete endpoints: %v\n", err)
	}

	// 7. Delete all entities for this project from database
	if err := s.entityRepo.HardDeleteByProjectID(project.ID); err != nil {
		fmt.Printf("Warning: failed to delete entities: %v\n", err)
	}

	// 8. Delete the project from database
	if err := s.projectRepo.HardDeleteByUUID(projectUUID); err != nil {
		return fmt.Errorf("failed to delete project from database: %w", err)
	}

	return nil
}

// dropServiceDatabase drops the database used by the generated service
func (s *DeploymentService) dropServiceDatabase(project *models.Project) error {
	// Skip if no database configured
	if project.DBHost == "" || project.DBName == "" {
		return nil
	}

	// Build DSN to connect to MySQL server (without specific database)
	// Use host.docker.internal for connecting from Lambra backend container to MySQL
	dbHost := project.DBHost
	if dbHost == "lambra-mysql" || dbHost == "mysql" {
		dbHost = "lambra-mysql" // Use Docker network name
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		project.DBUser,
		project.DBPassword,
		dbHost,
		project.DBPort,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL: %w", err)
	}

	// Drop the database
	dropSQL := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", project.DBName)
	if _, err := db.Exec(dropSQL); err != nil {
		return fmt.Errorf("failed to drop database %s: %w", project.DBName, err)
	}

	fmt.Printf("Successfully dropped database: %s\n", project.DBName)
	return nil
}

// createSnapshot creates a snapshot of the current project state before deployment
func (s *DeploymentService) createSnapshot(project *models.Project) (*models.GenerationSnapshot, error) {
	// Get all entities for the project
	entities, err := s.entityRepo.GetByProjectID(project.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	// Get all endpoints for each entity with entity name for mapping during rollback
	var allEndpoints []models.SnapshotEndpoint
	for _, entity := range entities {
		endpoints, err := s.endpointRepo.GetByEntityID(entity.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get endpoints for entity %s: %w", entity.Name, err)
		}
		// Wrap each endpoint with entity name for proper mapping during rollback
		for _, ep := range endpoints {
			allEndpoints = append(allEndpoints, models.SnapshotEndpoint{
				Endpoint:   ep,
				EntityName: entity.Name,
			})
		}
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
	// Use --force-recreate to avoid conflicts with existing containers
	cmd := exec.Command("docker", "compose", "up", "-d", "--build", "--force-recreate", "--remove-orphans")
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

// MigrationOptions contains options for database migrations during code generation
type MigrationOptions struct {
	ResetDatabase bool     // Drop all tables before creating
	TablesToDrop  []string // Specific tables to drop (deleted entities)
}

func (s *DeploymentService) generateGoFiles(project *models.Project, serviceDir string, port int, migOpts *MigrationOptions) error {
	if migOpts == nil {
		migOpts = &MigrationOptions{}
	}
	serviceName := toKebabCase(project.Name)

	// Get entities for this project
	entities, err := s.entityRepo.GetByProjectID(project.ID)
	if err != nil {
		return fmt.Errorf("failed to get entities: %w", err)
	}

	// Get relations for this project
	relations, err := s.relationRepo.GetByProject(project.ID)
	if err != nil {
		return fmt.Errorf("failed to get relations: %w", err)
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

	// Generate DROP TABLE statements if needed
	if migOpts.ResetDatabase {
		// Drop all current entity tables in reverse order (to handle FK constraints)
		for i := len(entities) - 1; i >= 0; i-- {
			tableName := entities[i].TableName
			migrations.WriteString(fmt.Sprintf("\t\t`DROP TABLE IF EXISTS %s`,\n", tableName))
		}
		// Also drop any junction tables
		for _, entity := range entities {
			junctionTables := s.getJunctionTableNames(entity, relations)
			for _, jt := range junctionTables {
				migrations.WriteString(fmt.Sprintf("\t\t`DROP TABLE IF EXISTS %s`,\n", jt))
			}
		}
	} else if len(migOpts.TablesToDrop) > 0 {
		// Drop specific tables (deleted entities)
		for _, tableName := range migOpts.TablesToDrop {
			migrations.WriteString(fmt.Sprintf("\t\t`DROP TABLE IF EXISTS %s`,\n", tableName))
		}
	}

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

		// Generate migration SQL for this entity (with relations)
		migrationSQL := s.generateMigrationSQL(entity, relations)
		migrations.WriteString(migrationSQL)

		// Generate junction tables for many-to-many relationships
		junctionSQL := s.generateJunctionTableSQL(entity, relations)
		migrations.WriteString(junctionSQL)

		// Get endpoints for this entity
		endpoints, err := s.endpointRepo.GetByEntityID(entity.ID)
		if err != nil {
			continue
		}

		// Generate routes for each endpoint
		// IMPORTANT: Handler method names must match the template-generated methods
		// which are based on entity name, not endpoint name.
		// Map HTTP method to handler method based on entity name.
		entityNamePlural := pluralize(entityNamePascal)
		for _, endpoint := range endpoints {
			// Derive handler method from HTTP method + entity name (not endpoint.Name)
			// This ensures routes call methods that actually exist in the generated handler
			handlerMethod := s.deriveHandlerMethod(endpoint.Method, endpoint.Path, entityNamePascal, entityNamePlural)
			routes.WriteString(fmt.Sprintf("\tr.%s(\"%s\", %sHandler.%s)\n",
				endpoint.Method, endpoint.Path, entityNameLower, handlerMethod))
		}

		// Add default CRUD routes if no endpoints exist (using query params per BSI UII rules)
		if len(endpoints) == 0 {
			basePath := "/" + entityNameSnake + "s"
			routes.WriteString(fmt.Sprintf("\tr.GET(\"%s\", %sHandler.List%s)\n", basePath, entityNameLower, entityNamePlural))
			routes.WriteString(fmt.Sprintf("\tr.GET(\"%s/detail\", %sHandler.Get%s)\n", basePath, entityNameLower, entityNamePascal))
			routes.WriteString(fmt.Sprintf("\tr.POST(\"%s\", %sHandler.Create%s)\n", basePath, entityNameLower, entityNamePascal))
			routes.WriteString(fmt.Sprintf("\tr.PUT(\"%s/update\", %sHandler.Update%s)\n", basePath, entityNameLower, entityNamePascal))
			routes.WriteString(fmt.Sprintf("\tr.DELETE(\"%s/delete\", %sHandler.Delete%s)\n", basePath, entityNameLower, entityNamePascal))
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
func (s *DeploymentService) generateMigrationSQL(entity models.Entity, relations []models.Relation) string {
	// Parse fields from JSON
	var fields []models.EntityField
	if err := json.Unmarshal(entity.Fields, &fields); err != nil {
		return ""
	}

	var sql strings.Builder
	var foreignKeys []string
	tableName := entity.TableName

	sql.WriteString(fmt.Sprintf("\t\t`CREATE TABLE IF NOT EXISTS %s (\n", tableName))
	sql.WriteString("\t\t\tid BIGINT PRIMARY KEY AUTO_INCREMENT,\n")
	sql.WriteString("\t\t\tuuid CHAR(36) NOT NULL UNIQUE,\n")

	// 1. Add regular fields from entity.Fields
	// Track FK fields to avoid duplicates
	fkFields := make(map[string]bool)
	
	// First pass: collect FK field names from relations
	for _, rel := range relations {
		if rel.RelationType == models.RelationTypeBelongsTo && rel.SourceEntityID == entity.ID {
			if rel.SourceFieldName != "" {
				fkFields[rel.SourceFieldName] = true
			}
		} else if (rel.RelationType == models.RelationTypeHasOne || rel.RelationType == models.RelationTypeHasMany) && rel.TargetEntityID == entity.ID {
			if rel.SourceFieldName != "" {
				fkFields[rel.SourceFieldName] = true
			}
		}
	}
	
	// Second pass: add regular fields (skip if it's a FK field)
	for _, field := range fields {
		// Skip old relation fields (they should use relations table now)
		if field.Type == "relation" {
			continue
		}
		
		// Skip if this field is a FK from a relation
		fieldName := toSnakeCase(field.Name)
		if fkFields[fieldName] {
			continue // Skip duplicate FK field
		}

		// Regular fields
		columnName := toSnakeCase(field.Name)
		columnType := s.getSQLType(field.Type, field.Length)
		notNull := ""
		if field.Required {
			notNull = " NOT NULL"
		}
		unique := ""
		if field.Unique {
			unique = " UNIQUE"
		}
		defaultVal := ""
		if field.DefaultValue != "" {
			defaultVal = fmt.Sprintf(" DEFAULT %s", field.DefaultValue)
		}
		sql.WriteString(fmt.Sprintf("\t\t\t%s %s%s%s%s,\n", columnName, columnType, notNull, unique, defaultVal))
	}

	// 2. Add FK columns from relations table
	for _, rel := range relations {
		shouldAddFK := false
		fkColumn := ""
		targetTable := ""
		var targetEntityID int64
		
		// Determine if this entity should have FK column based on relation type
		if rel.RelationType == models.RelationTypeBelongsTo && rel.SourceEntityID == entity.ID {
			// belongsTo: source has FK to target
			shouldAddFK = true
			fkColumn = rel.SourceFieldName
			if fkColumn == "" {
				targetEntity, _ := s.entityRepo.GetByID(rel.TargetEntityID)
				if targetEntity != nil {
					fkColumn = toSnakeCase(targetEntity.Name) + "_id"
				}
			}
			targetEntityID = rel.TargetEntityID
		} else if (rel.RelationType == models.RelationTypeHasOne || rel.RelationType == models.RelationTypeHasMany) && rel.TargetEntityID == entity.ID {
			// hasOne/hasMany: target has FK to source
			shouldAddFK = true
			fkColumn = rel.SourceFieldName
			if fkColumn == "" {
				sourceEntity, _ := s.entityRepo.GetByID(rel.SourceEntityID)
				if sourceEntity != nil {
					fkColumn = toSnakeCase(sourceEntity.Name) + "_id"
				}
			}
			targetEntityID = rel.SourceEntityID
		}
		
		if shouldAddFK && fkColumn != "" {
			notNull := ""
			if rel.Required {
				notNull = " NOT NULL"
			}

			sql.WriteString(fmt.Sprintf("\t\t\t%s BIGINT%s,\n", fkColumn, notNull))

			// Add FK constraint
			onDelete := rel.OnDelete
			if onDelete == "" {
				onDelete = "CASCADE"
			}
			onUpdate := rel.OnUpdate
			if onUpdate == "" {
				onUpdate = "CASCADE"
			}
			
			// Get target table name
			if s.entityRepo != nil {
				targetEntity, _ := s.entityRepo.GetByID(targetEntityID)
				if targetEntity != nil {
					targetTable = targetEntity.TableName
				}
			}
			
			// Fallback: derive table name from entity name if repo not available (testing)
			if targetTable == "" {
				if rel.RelationType == models.RelationTypeBelongsTo {
					targetTable = toSnakeCase(pluralize(rel.TargetEntityName))
				} else {
					targetTable = toSnakeCase(pluralize(rel.SourceEntityName))
				}
			}
			
			if targetTable != "" {
				foreignKeys = append(foreignKeys, fmt.Sprintf("\t\t\tFOREIGN KEY (%s) REFERENCES %s(id) ON DELETE %s ON UPDATE %s", fkColumn, targetTable, onDelete, onUpdate))
			}
		}
	}

	sql.WriteString("\t\t\tcreated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,\n")
	sql.WriteString("\t\t\tupdated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n")
	sql.WriteString("\t\t\tdeleted_at TIMESTAMP NULL DEFAULT NULL,\n")
	sql.WriteString(fmt.Sprintf("\t\t\tINDEX idx_%s_uuid (uuid),\n", tableName))
	sql.WriteString(fmt.Sprintf("\t\t\tINDEX idx_%s_deleted_at (deleted_at)", tableName))

	// Add FK constraints
	for _, fk := range foreignKeys {
		sql.WriteString(",\n")
		sql.WriteString(fk)
	}

	sql.WriteString("\n\t\t)`,\n")

	return sql.String()
}

// generateJunctionTableSQL generates CREATE TABLE for many-to-many relationships
func (s *DeploymentService) generateJunctionTableSQL(entity models.Entity, relations []models.Relation) string {
	var sql strings.Builder

	// Find manyToMany relations where this entity is the source
	for _, rel := range relations {
		if rel.SourceEntityID == entity.ID && rel.RelationType == models.RelationTypeManyToMany {
			// Get source and target entities
			sourceEntity := entity
			targetEntity, err := s.entityRepo.GetByID(rel.TargetEntityID)
			if err != nil || targetEntity == nil {
				continue
			}

			// Use junction table name from relation (or generate if empty)
			junctionTableName := rel.JunctionTable.String
			if !rel.JunctionTable.Valid || junctionTableName == "" {
				// Generate junction table name (alphabetical order)
				sourceTable := sourceEntity.TableName
				targetTable := targetEntity.TableName
				if sourceTable < targetTable {
					junctionTableName = sourceTable + "_" + targetTable
				} else {
					junctionTableName = targetTable + "_" + sourceTable
				}
			}

			// Generate FK column names
			sourceFK := toSnakeCase(strings.TrimSuffix(sourceEntity.TableName, "s")) + "_id"
			targetFK := toSnakeCase(strings.TrimSuffix(targetEntity.TableName, "s")) + "_id"

			sql.WriteString(fmt.Sprintf("\t\t`CREATE TABLE IF NOT EXISTS %s (\n", junctionTableName))
			sql.WriteString(fmt.Sprintf("\t\t\t%s BIGINT NOT NULL,\n", sourceFK))
			sql.WriteString(fmt.Sprintf("\t\t\t%s BIGINT NOT NULL,\n", targetFK))
			sql.WriteString("\t\t\tcreated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,\n")
			sql.WriteString(fmt.Sprintf("\t\t\tPRIMARY KEY (%s, %s),\n", sourceFK, targetFK))
			sql.WriteString(fmt.Sprintf("\t\t\tFOREIGN KEY (%s) REFERENCES %s(id) ON DELETE CASCADE,\n", sourceFK, sourceEntity.TableName))
			sql.WriteString(fmt.Sprintf("\t\t\tFOREIGN KEY (%s) REFERENCES %s(id) ON DELETE CASCADE\n", targetFK, targetEntity.TableName))
			sql.WriteString("\t\t)`,\n")
		}
	}

	return sql.String()
}

// getJunctionTableNames returns junction table names for many-to-many relations in an entity
func (s *DeploymentService) getJunctionTableNames(entity models.Entity, relations []models.Relation) []string {
	var tables []string

	for _, rel := range relations {
		if rel.SourceEntityID == entity.ID && rel.RelationType == models.RelationTypeManyToMany {
			// Use junction table name from relation (or generate if empty)
			junctionTableName := rel.JunctionTable.String
			if !rel.JunctionTable.Valid || junctionTableName == "" {
				targetEntity, _ := s.entityRepo.GetByID(rel.TargetEntityID)
				if targetEntity != nil {
					sourceTable := entity.TableName
					targetTable := targetEntity.TableName
					if sourceTable < targetTable {
						junctionTableName = sourceTable + "_" + targetTable
					} else {
						junctionTableName = targetTable + "_" + sourceTable
					}
				}
			}
			if junctionTableName != "" {
				tables = append(tables, junctionTableName)
			}
		}
	}

	return tables
}

// detectDeletedTables compares previous snapshot with current entities to find deleted tables
func (s *DeploymentService) detectDeletedTables(previousSnapshot *models.GenerationSnapshot, projectID int64) []string {
	var tablesToDrop []string

	// Parse previous snapshot metadata
	var previousMetadata models.SnapshotMetadata
	if err := json.Unmarshal(previousSnapshot.Metadata, &previousMetadata); err != nil {
		return nil
	}

	// Get current entities
	currentEntities, err := s.entityRepo.GetByProjectID(projectID)
	if err != nil {
		return nil
	}

	// Get current relations
	currentRelations, err := s.relationRepo.GetByProject(projectID)
	if err != nil {
		currentRelations = []models.Relation{} // Continue without relations if error
	}

	// Create map of current table names
	currentTables := make(map[string]bool)
	for _, entity := range currentEntities {
		currentTables[entity.TableName] = true
		// Also add junction tables
		for _, jt := range s.getJunctionTableNames(entity, currentRelations) {
			currentTables[jt] = true
		}
	}

	// Find tables that existed in previous snapshot but not in current entities
	for _, prevEntity := range previousMetadata.Entities {
		if !currentTables[prevEntity.TableName] {
			tablesToDrop = append(tablesToDrop, prevEntity.TableName)
		}
		// Check for junction tables from previous entities (using prev relations)
		for _, jt := range s.getJunctionTableNames(prevEntity, previousMetadata.Relations) {
			if !currentTables[jt] {
				tablesToDrop = append(tablesToDrop, jt)
			}
		}
	}

	return tablesToDrop
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

// GetDeploymentsByProjectUUID retrieves all deployments for a project
func (s *DeploymentService) GetDeploymentsByProjectUUID(ctx context.Context, projectUUID string, limit, offset int) ([]models.Deployment, int64, error) {
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get project: %w", err)
	}

	deployments, total, err := s.deploymentRepo.GetByProjectID(project.ID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deployments: %w", err)
	}

	return deployments, total, nil
}

// GetDeploymentByUUID retrieves a single deployment by UUID
func (s *DeploymentService) GetDeploymentByUUID(ctx context.Context, deploymentUUID string) (*models.DeploymentWithLogs, error) {
	deployment, err := s.deploymentRepo.GetByUUID(deploymentUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	logs, err := s.deploymentLogRepo.GetByDeploymentID(deployment.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment logs: %w", err)
	}

	return &models.DeploymentWithLogs{
		Deployment: *deployment,
		Logs:       logs,
	}, nil
}

// GetDeploymentLogs retrieves logs for a deployment
func (s *DeploymentService) GetDeploymentLogs(ctx context.Context, deploymentUUID string, limit, offset int) ([]models.DeploymentLog, int64, error) {
	deployment, err := s.deploymentRepo.GetByUUID(deploymentUUID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deployment: %w", err)
	}

	logs, total, err := s.deploymentLogRepo.GetByDeploymentIDPaginated(deployment.ID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deployment logs: %w", err)
	}

	return logs, total, nil
}

// GetDeploymentLogsSince retrieves logs created after a specific time
func (s *DeploymentService) GetDeploymentLogsSince(ctx context.Context, deploymentUUID string, since time.Time) ([]models.DeploymentLog, error) {
	deployment, err := s.deploymentRepo.GetByUUID(deploymentUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	logs, err := s.deploymentLogRepo.GetByDeploymentIDSince(deployment.ID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment logs: %w", err)
	}

	return logs, nil
}

// GetDeploymentLogsAfterID retrieves logs with ID greater than specified (for SSE streaming)
func (s *DeploymentService) GetDeploymentLogsAfterID(ctx context.Context, deploymentUUID string, afterID int64) ([]models.DeploymentLog, error) {
	deployment, err := s.deploymentRepo.GetByUUID(deploymentUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	logs, err := s.deploymentLogRepo.GetByDeploymentIDAfterId(deployment.ID, afterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment logs: %w", err)
	}

	return logs, nil
}

// GetLatestDeploymentByProjectUUID retrieves the latest deployment for a project
func (s *DeploymentService) GetLatestDeploymentByProjectUUID(ctx context.Context, projectUUID string) (*models.DeploymentWithLogs, error) {
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	deployment, err := s.deploymentRepo.GetLatestByProjectID(project.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest deployment: %w", err)
	}
	if deployment == nil {
		return nil, nil
	}

	logs, err := s.deploymentLogRepo.GetByDeploymentID(deployment.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment logs: %w", err)
	}

	return &models.DeploymentWithLogs{
		Deployment: *deployment,
		Logs:       logs,
	}, nil
}

// GetContainerLogs retrieves Docker container logs for a running service
func (s *DeploymentService) GetContainerLogs(ctx context.Context, projectUUID string, tail int) (string, error) {
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return "", fmt.Errorf("failed to get project: %w", err)
	}

	serviceName := toKebabCase(project.Name)

	// Check if container is running
	status := s.checkContainerStatus(serviceName)
	if status != "running" {
		return "", fmt.Errorf("container is not running")
	}

	// Get container logs
	args := []string{"logs", serviceName}
	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}

	return string(output), nil
}

// StreamContainerLogs creates a command that streams container logs (for SSE)
func (s *DeploymentService) StreamContainerLogs(ctx context.Context, projectUUID string) (*exec.Cmd, error) {
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	serviceName := toKebabCase(project.Name)

	// Check if container is running
	status := s.checkContainerStatus(serviceName)
	if status != "running" {
		return nil, fmt.Errorf("container is not running")
	}

	// Create command to stream logs
	cmd := exec.CommandContext(ctx, "docker", "logs", "-f", "--tail", "100", serviceName)
	return cmd, nil
}

// deriveHandlerMethod derives the handler method name from HTTP method and entity name
// This ensures routes call methods that actually exist in the generated handler template
// Handler template generates: List{Plural}, Get{Entity}, Create{Entity}, Update{Entity}, Delete{Entity}
func (s *DeploymentService) deriveHandlerMethod(httpMethod, path, entityName, entityNamePlural string) string {
	switch httpMethod {
	case "GET":
		// Check if it's a list or single item endpoint
		// List: /entities, /entities/
		// Detail: /entities/detail, /entities/:id, /entities/{id}
		if strings.Contains(path, "/detail") || strings.Contains(path, "/:") || strings.Contains(path, "/{") {
			return "Get" + entityName
		}
		return "List" + entityNamePlural
	case "POST":
		return "Create" + entityName
	case "PUT", "PATCH":
		return "Update" + entityName
	case "DELETE":
		return "Delete" + entityName
	default:
		// Fallback to entity-based name
		return "Handle" + entityName
	}
}
