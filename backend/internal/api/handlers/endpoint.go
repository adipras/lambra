package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/service"
	"github.com/yourusername/lambra/pkg/response"
)

type EndpointHandler struct {
	service *service.EndpointService
}

func NewEndpointHandler(service *service.EndpointService) *EndpointHandler {
	return &EndpointHandler{service: service}
}

// CreateEndpoint creates a new endpoint for an entity
func (h *EndpointHandler) CreateEndpoint(c *gin.Context) {
	var req models.CreateEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	endpoint, err := h.service.CreateEndpoint(&req)
	if err != nil {
		response.InternalError(c, "Failed to create endpoint", err)
		return
	}

	response.Created(c, endpoint, "Endpoint created successfully")
}

// GetEndpoint retrieves an endpoint by UUID
func (h *EndpointHandler) GetEndpoint(c *gin.Context) {
	uuid := c.Param("id")
	if uuid == "" {
		response.BadRequest(c, "Invalid endpoint ID", nil)
		return
	}

	endpoint, err := h.service.GetEndpointByUUID(uuid)
	if err != nil {
		response.NotFound(c, "Endpoint not found")
		return
	}

	response.Success(c, endpoint, "Endpoint retrieved successfully")
}

// GetEndpointsByProject retrieves all endpoints for a project
// GET /api/v1/projects/:id/endpoints
func (h *EndpointHandler) GetEndpointsByProject(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	endpoints, err := h.service.GetEndpointsByProjectUUID(projectID)
	if err != nil {
		response.InternalError(c, "Failed to retrieve endpoints", err)
		return
	}

	response.Success(c, endpoints, "Endpoints retrieved successfully")
}

// GetEndpointsByEntity retrieves all endpoints for an entity
// GET /api/v1/entities/:id/endpoints
func (h *EndpointHandler) GetEndpointsByEntity(c *gin.Context) {
	entityID := c.Param("id")
	if entityID == "" {
		response.BadRequest(c, "Invalid entity ID", nil)
		return
	}

	endpoints, err := h.service.GetEndpointsByEntityUUID(entityID)
	if err != nil {
		response.InternalError(c, "Failed to retrieve endpoints", err)
		return
	}

	response.Success(c, endpoints, "Endpoints retrieved successfully")
}

// UpdateEndpoint updates an endpoint by UUID
func (h *EndpointHandler) UpdateEndpoint(c *gin.Context) {
	uuid := c.Param("id")
	if uuid == "" {
		response.BadRequest(c, "Invalid endpoint ID", nil)
		return
	}

	var req models.UpdateEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	endpoint, err := h.service.UpdateEndpoint(uuid, &req)
	if err != nil {
		response.InternalError(c, "Failed to update endpoint", err)
		return
	}

	response.Success(c, endpoint, "Endpoint updated successfully")
}

// DeleteEndpoint deletes an endpoint by UUID (soft delete)
func (h *EndpointHandler) DeleteEndpoint(c *gin.Context) {
	uuid := c.Param("id")
	if uuid == "" {
		response.BadRequest(c, "Invalid endpoint ID", nil)
		return
	}

	err := h.service.DeleteEndpoint(uuid)
	if err != nil {
		response.InternalError(c, "Failed to delete endpoint", err)
		return
	}

	response.Success(c, nil, "Endpoint deleted successfully")
}
