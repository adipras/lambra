package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/service"
	"github.com/yourusername/lambra/pkg/response"
)

type EntityHandler struct {
	service *service.EntityService
}

func NewEntityHandler(service *service.EntityService) *EntityHandler {
	return &EntityHandler{service: service}
}

// CreateEntity creates a new entity for a project
// POST /api/v1/projects/:id/entities
func (h *EntityHandler) CreateEntity(c *gin.Context) {
	var req models.CreateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	// Get project id from URL param and set it in request
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}
	req.ProjectUUID = projectID

	entity, err := h.service.CreateEntity(&req)
	if err != nil {
		response.InternalError(c, "Failed to create entity", err)
		return
	}

	response.Created(c, entity, "Entity created successfully")
}

// GetEntity retrieves an entity by UUID
func (h *EntityHandler) GetEntity(c *gin.Context) {
	uuid := c.Param("id")
	if uuid == "" {
		response.BadRequest(c, "Invalid entity ID", nil)
		return
	}

	entity, err := h.service.GetEntityByUUID(uuid)
	if err != nil {
		response.NotFound(c, "Entity not found")
		return
	}

	response.Success(c, entity, "Entity retrieved successfully")
}

// GetEntitiesByProject retrieves all entities for a project
// GET /api/v1/projects/:id/entities
func (h *EntityHandler) GetEntitiesByProject(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	entities, err := h.service.GetEntitiesByProjectUUID(projectID)
	if err != nil {
		response.InternalError(c, "Failed to retrieve entities", err)
		return
	}

	response.Success(c, entities, "Entities retrieved successfully")
}

// UpdateEntity updates an entity by UUID
func (h *EntityHandler) UpdateEntity(c *gin.Context) {
	uuid := c.Param("id")
	if uuid == "" {
		response.BadRequest(c, "Invalid entity ID", nil)
		return
	}

	var req models.UpdateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	entity, err := h.service.UpdateEntity(uuid, &req)
	if err != nil {
		response.InternalError(c, "Failed to update entity", err)
		return
	}

	response.Success(c, entity, "Entity updated successfully")
}

// DeleteEntity deletes an entity by UUID (soft delete)
func (h *EntityHandler) DeleteEntity(c *gin.Context) {
	uuid := c.Param("id")
	if uuid == "" {
		response.BadRequest(c, "Invalid entity ID", nil)
		return
	}

	err := h.service.DeleteEntity(uuid)
	if err != nil {
		response.InternalError(c, "Failed to delete entity", err)
		return
	}

	response.Success(c, nil, "Entity deleted successfully")
}
