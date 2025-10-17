package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/lambra/internal/service"
)

// GeneratorHandler handles code generation HTTP requests
type GeneratorHandler struct {
	service *service.GeneratorService
}

// NewGeneratorHandler creates a new generator handler
func NewGeneratorHandler(service *service.GeneratorService) *GeneratorHandler {
	return &GeneratorHandler{
		service: service,
	}
}

// GenerateEntityRequest represents a request to generate code for an entity
type GenerateEntityRequest struct {
	EntityID  int64  `json:"entity_id" binding:"required"`
	OutputDir string `json:"output_dir"`
}

// GenerateProjectRequest represents a request to generate code for a project
type GenerateProjectRequest struct {
	ProjectID int64  `json:"project_id" binding:"required"`
	OutputDir string `json:"output_dir"`
}

// GenerateEntity generates code for a specific entity
// @Summary Generate code for an entity
// @Description Generates model, repository, service, handler, DTO, and migration files for an entity
// @Tags generator
// @Accept json
// @Produce json
// @Param request body GenerateEntityRequest true "Generation request"
// @Success 200 {object} service.GenerateCodeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/generate/entity [post]
func (h *GeneratorHandler) GenerateEntity(c *gin.Context) {
	var req GenerateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.OutputDir == "" {
		req.OutputDir = "./generated"
	}

	response, err := h.service.GenerateEntity(c.Request.Context(), req.EntityID, req.OutputDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GenerateProject generates code for all entities in a project
// @Summary Generate code for a project
// @Description Generates code for all entities in a project
// @Tags generator
// @Accept json
// @Produce json
// @Param request body GenerateProjectRequest true "Generation request"
// @Success 200 {object} service.GenerateCodeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/generate/project [post]
func (h *GeneratorHandler) GenerateProject(c *gin.Context) {
	var req GenerateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.OutputDir == "" {
		req.OutputDir = "./generated"
	}

	response, err := h.service.GenerateProject(c.Request.Context(), req.ProjectID, req.OutputDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// PreviewEntity previews generated code without writing files
// @Summary Preview generated code
// @Description Returns the generated code preview without writing to filesystem
// @Tags generator
// @Accept json
// @Produce json
// @Param id path int true "Entity ID"
// @Success 200 {object} service.GenerateCodeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/generate/preview/:id [get]
func (h *GeneratorHandler) PreviewEntity(c *gin.Context) {
	entityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	response, err := h.service.PreviewEntity(c.Request.Context(), entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetGeneratedFilesList returns list of files that will be generated
// @Summary Get list of files to be generated
// @Description Returns the list of files that will be generated for an entity
// @Tags generator
// @Accept json
// @Produce json
// @Param id path int true "Entity ID"
// @Success 200 {object} map[string][]string
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/generate/files/:id [get]
func (h *GeneratorHandler) GetGeneratedFilesList(c *gin.Context) {
	entityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	files, err := h.service.GetGeneratedFilesList(c.Request.Context(), entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
