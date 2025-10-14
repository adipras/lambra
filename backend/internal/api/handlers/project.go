package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/service"
	"github.com/yourusername/lambra/pkg/response"
)

type ProjectHandler struct {
	service *service.ProjectService
}

func NewProjectHandler(service *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	project, err := h.service.CreateProject(&req)
	if err != nil {
		response.InternalError(c, "Failed to create project", err)
		return
	}

	response.Created(c, project, "Project created successfully")
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	project, err := h.service.GetProjectWithRelations(id)
	if err != nil {
		response.NotFound(c, "Project not found")
		return
	}

	response.Success(c, project, "Project retrieved successfully")
}

func (h *ProjectHandler) GetAllProjects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	projects, total, err := h.service.GetAllProjects(page, limit)
	if err != nil {
		response.InternalError(c, "Failed to retrieve projects", err)
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	pagination := response.Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	}

	response.SuccessWithPagination(c, projects, pagination, "Projects retrieved successfully")
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	project, err := h.service.UpdateProject(id, &req)
	if err != nil {
		response.InternalError(c, "Failed to update project", err)
		return
	}

	response.Success(c, project, "Project updated successfully")
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	err := h.service.DeleteProject(id)
	if err != nil {
		response.InternalError(c, "Failed to delete project", err)
		return
	}

	response.Success(c, nil, "Project deleted successfully")
}
