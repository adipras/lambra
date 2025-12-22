package handlers

import (
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
