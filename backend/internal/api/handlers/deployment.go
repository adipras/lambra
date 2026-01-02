package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/lambra/internal/service"
	"github.com/yourusername/lambra/pkg/response"
)

type DeploymentHandler struct {
	service *service.DeploymentService
}

func NewDeploymentHandler(service *service.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{service: service}
}

// DeployProject deploys a project as a Docker service
// POST /api/v1/projects/:id/deploy
func (h *DeploymentHandler) DeployProject(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	result, err := h.service.DeployProject(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c, "Failed to deploy project", err)
		return
	}

	response.Success(c, result, "Project deployed successfully")
}

// StartService starts a deployed service
// POST /api/v1/projects/:id/start
func (h *DeploymentHandler) StartService(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	result, err := h.service.StartService(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c, "Failed to start service", err)
		return
	}

	response.Success(c, result, "Service started successfully")
}

// StopService stops a running service
// POST /api/v1/projects/:id/stop
func (h *DeploymentHandler) StopService(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	result, err := h.service.StopService(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c, "Failed to stop service", err)
		return
	}

	response.Success(c, result, "Service stopped successfully")
}

// GetServiceStatus gets the status of a deployed service
// GET /api/v1/projects/:id/status
func (h *DeploymentHandler) GetServiceStatus(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	result, err := h.service.GetServiceStatus(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c, "Failed to get service status", err)
		return
	}

	response.Success(c, result, "Service status retrieved successfully")
}

// RedeployService redeploys a service (down + up for cache clearing)
// POST /api/v1/projects/:id/redeploy
func (h *DeploymentHandler) RedeployService(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	result, err := h.service.RedeployService(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c, "Failed to redeploy service", err)
		return
	}

	response.Success(c, result, "Service redeployed successfully")
}

// DestroyService destroys a service completely (stop containers, remove volumes, delete workspace)
// DELETE /api/v1/projects/:id/destroy
func (h *DeploymentHandler) DestroyService(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	err := h.service.DestroyServiceCompletely(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c, "Failed to destroy service", err)
		return
	}

	response.Success(c, nil, "Service destroyed successfully")
}

// GetProjectDeployments retrieves all deployments for a project
// GET /api/v1/projects/:id/deployments
func (h *DeploymentHandler) GetProjectDeployments(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	deployments, total, err := h.service.GetDeploymentsByProjectUUID(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to get deployments", err)
		return
	}

	response.SuccessWithMeta(c, deployments, "Deployments retrieved successfully", map[string]interface{}{
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetLatestDeployment retrieves the latest deployment for a project
// GET /api/v1/projects/:id/deployments/latest
func (h *DeploymentHandler) GetLatestDeployment(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	deployment, err := h.service.GetLatestDeploymentByProjectUUID(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c, "Failed to get latest deployment", err)
		return
	}

	if deployment == nil {
		response.NotFound(c, "No deployments found")
		return
	}

	response.Success(c, deployment, "Latest deployment retrieved successfully")
}

// GetDeployment retrieves a single deployment by UUID
// GET /api/v1/deployments/:id
func (h *DeploymentHandler) GetDeployment(c *gin.Context) {
	deploymentID := c.Param("id")
	if deploymentID == "" {
		response.BadRequest(c, "Invalid deployment ID", nil)
		return
	}

	deployment, err := h.service.GetDeploymentByUUID(c.Request.Context(), deploymentID)
	if err != nil {
		response.InternalError(c, "Failed to get deployment", err)
		return
	}

	response.Success(c, deployment, "Deployment retrieved successfully")
}

// GetDeploymentLogs retrieves logs for a deployment
// GET /api/v1/deployments/:id/logs
func (h *DeploymentHandler) GetDeploymentLogs(c *gin.Context) {
	deploymentID := c.Param("id")
	if deploymentID == "" {
		response.BadRequest(c, "Invalid deployment ID", nil)
		return
	}

	limit := 100
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	logs, total, err := h.service.GetDeploymentLogs(c.Request.Context(), deploymentID, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to get deployment logs", err)
		return
	}

	response.SuccessWithMeta(c, logs, "Deployment logs retrieved successfully", map[string]interface{}{
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// StreamDeploymentLogs streams deployment logs via SSE
// GET /api/v1/deployments/:id/logs/stream
func (h *DeploymentHandler) StreamDeploymentLogs(c *gin.Context) {
	deploymentID := c.Param("id")
	if deploymentID == "" {
		c.JSON(400, gin.H{"error": "Invalid deployment ID"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Get initial logs
	var lastID int64 = 0
	logs, _, err := h.service.GetDeploymentLogs(c.Request.Context(), deploymentID, 100, 0)
	if err == nil && len(logs) > 0 {
		for _, log := range logs {
			data, _ := json.Marshal(log)
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
			lastID = log.ID
		}
	}

	// Stream new logs
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			newLogs, err := h.service.GetDeploymentLogsAfterID(c.Request.Context(), deploymentID, lastID)
			if err != nil {
				continue
			}
			for _, log := range newLogs {
				data, _ := json.Marshal(log)
				fmt.Fprintf(c.Writer, "data: %s\n\n", data)
				c.Writer.Flush()
				lastID = log.ID
			}
		}
	}
}

// GetContainerLogs retrieves Docker container logs
// GET /api/v1/projects/:id/container-logs
func (h *DeploymentHandler) GetContainerLogs(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	tail := 100
	if t := c.Query("tail"); t != "" {
		if parsed, err := strconv.Atoi(t); err == nil {
			tail = parsed
		}
	}

	logs, err := h.service.GetContainerLogs(c.Request.Context(), projectID, tail)
	if err != nil {
		response.InternalError(c, "Failed to get container logs", err)
		return
	}

	response.Success(c, gin.H{"logs": logs}, "Container logs retrieved successfully")
}

// StreamContainerLogs streams Docker container logs via SSE
// GET /api/v1/projects/:id/container-logs/stream
func (h *DeploymentHandler) StreamContainerLogs(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(400, gin.H{"error": "Invalid project ID"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Create stream command
	cmd, err := h.service.StreamContainerLogs(c.Request.Context(), projectID)
	if err != nil {
		fmt.Fprintf(c.Writer, "data: {\"error\": \"%s\"}\n\n", err.Error())
		c.Writer.Flush()
		return
	}

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(c.Writer, "data: {\"error\": \"Failed to get stdout pipe\"}\n\n", err.Error())
		c.Writer.Flush()
		return
	}

	// Get stderr pipe
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(c.Writer, "data: {\"error\": \"Failed to get stderr pipe\"}\n\n", err.Error())
		c.Writer.Flush()
		return
	}

	// Start command
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(c.Writer, "data: {\"error\": \"Failed to start log stream\"}\n\n", err.Error())
		c.Writer.Flush()
		return
	}

	// Stream stdout and stderr concurrently
	done := make(chan bool)

	streamOutput := func(reader io.Reader, stream string) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			data := map[string]string{
				"stream":    stream,
				"message":   line,
				"timestamp": time.Now().Format(time.RFC3339),
			}
			jsonData, _ := json.Marshal(data)
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
			c.Writer.Flush()
		}
	}

	go func() {
		streamOutput(stdout, "stdout")
		done <- true
	}()

	go func() {
		streamOutput(stderr, "stderr")
		done <- true
	}()

	// Wait for context cancellation or command completion
	select {
	case <-c.Request.Context().Done():
		cmd.Process.Kill()
	case <-done:
		// One stream finished
	}

	cmd.Wait()
}
