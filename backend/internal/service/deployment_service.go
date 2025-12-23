package service

import (
	"bytes"
	"context"
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
	generatorSvc   *GeneratorService
	workspacePath  string
	basePort       int
}

// NewDeploymentService creates a new deployment service
func NewDeploymentService(
	projectRepo *repository.ProjectRepository,
	entityRepo *repository.EntityRepository,
	endpointRepo *repository.EndpointRepository,
	generatorSvc *GeneratorService,
	workspacePath string,
) *DeploymentService {
	return &DeploymentService{
		projectRepo:   projectRepo,
		entityRepo:    entityRepo,
		endpointRepo:  endpointRepo,
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
	URL         string `json:"url,omitempty"`
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
}

// DeployProject generates code and deploys the service
func (s *DeploymentService) DeployProject(ctx context.Context, projectUUID string) (*DeployResult, error) {
	// Get project
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	serviceName := toKebabCase(project.Name)
	serviceDir := filepath.Join(s.workspacePath, serviceName)

	// Create service directory
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create service directory: %w", err)
	}

	// Generate code
	response, err := s.generatorSvc.GenerateProjectByUUID(ctx, projectUUID, serviceDir)
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
	if err := s.startService(serviceDir, serviceName); err != nil {
		return nil, fmt.Errorf("failed to start service: %w", err)
	}

	return &DeployResult{
		Success:     true,
		Message:     fmt.Sprintf("Service %s deployed successfully", serviceName),
		ServiceName: serviceName,
		Port:        port,
		URL:         fmt.Sprintf("http://localhost:%d", port),
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

	if err := s.startService(serviceDir, serviceName); err != nil {
		return nil, fmt.Errorf("failed to start service: %w", err)
	}

	port := s.getPortForProject(project.ID)
	return &ServiceStatus{
		ProjectID:   projectUUID,
		ServiceName: serviceName,
		Status:      "running",
		Port:        port,
		URL:         fmt.Sprintf("http://localhost:%d", port),
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
	}

	return result, nil
}

// Helper methods

func (s *DeploymentService) getPortForProject(projectID int64) int {
	// Simple port assignment based on project ID
	return s.basePort + int(projectID%1000)
}

func (s *DeploymentService) startService(serviceDir, serviceName string) error {
	cmd := exec.Command("docker", "compose", "up", "-d", "--build")
	cmd.Dir = serviceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *DeploymentService) stopService(serviceDir string) error {
	cmd := exec.Command("docker", "compose", "down")
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
	dbPort := 3400 + int(project.ID%100) // Unique DB port

	data := map[string]interface{}{
		"ServiceName":      serviceName,
		"Port":             port,
		"DatabaseName":     strings.ReplaceAll(serviceName, "-", "_") + "_db",
		"DatabaseUser":     serviceName,
		"DatabasePassword": serviceName + "_secret",
		"DatabasePort":     dbPort,
		"Environment":      "development",
		"GinMode":          "debug",
	}

	// Generate docker-compose.yml
	dockerComposeTmpl := `services:
  {{.ServiceName}}-db:
    image: mysql:8.0
    container_name: {{.ServiceName}}-db
    environment:
      MYSQL_ROOT_PASSWORD: root_password
      MYSQL_DATABASE: {{.DatabaseName}}
      MYSQL_USER: {{.DatabaseUser}}
      MYSQL_PASSWORD: {{.DatabasePassword}}
    ports:
      - "{{.DatabasePort}}:3306"
    volumes:
      - {{.ServiceName}}_db_data:/var/lib/mysql
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-u", "root", "-proot_password"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - {{.ServiceName}}-network

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
      DB_HOST: {{.ServiceName}}-db
      DB_PORT: 3306
      DB_USER: {{.DatabaseUser}}
      DB_PASSWORD: {{.DatabasePassword}}
      DB_NAME: {{.DatabaseName}}
    depends_on:
      {{.ServiceName}}-db:
        condition: service_healthy
    networks:
      - {{.ServiceName}}-network
    restart: unless-stopped

volumes:
  {{.ServiceName}}_db_data:

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

	// main.go
	mainGoContent := fmt.Sprintf(`package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

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
		log.Printf("Warning: Could not connect to database: %%v", err)
	} else {
		defer db.Close()
		log.Println("Database connected successfully")
	}

	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "%s",
		})
	})

	// TODO: Add generated handlers here

	log.Printf("Server starting on port %%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
`, port, serviceName)

	if err := os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(mainGoContent), 0644); err != nil {
		return err
	}

	// Run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = serviceDir
	cmd.Run() // Ignore error, it might fail without network

	return nil
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
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result = append(result, '-')
			}
			result = append(result, r+32) // Convert to lowercase
		} else if r == ' ' || r == '_' {
			result = append(result, '-')
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
