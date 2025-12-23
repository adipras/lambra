package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/lambra/internal/service"
	"github.com/yourusername/lambra/pkg/response"
)

type ExportHandler struct {
	service *service.ExportService
}

func NewExportHandler(service *service.ExportService) *ExportHandler {
	return &ExportHandler{service: service}
}

// ExportOpenAPI generates OpenAPI 3.0 specification for a project
// GET /api/v1/projects/:id/export/openapi
func (h *ExportHandler) ExportOpenAPI(c *gin.Context) {
	projectUUID := c.Param("id")
	if projectUUID == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	spec, err := h.service.GenerateOpenAPISpec(projectUUID)
	if err != nil {
		response.InternalError(c, "Failed to generate OpenAPI spec", err)
		return
	}

	// Check if download is requested
	if c.Query("download") == "true" {
		jsonBytes, err := json.MarshalIndent(spec, "", "  ")
		if err != nil {
			response.InternalError(c, "Failed to serialize OpenAPI spec", err)
			return
		}

		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="openapi-%s.json"`, projectUUID[:8]))
		c.Header("Content-Type", "application/json")
		c.Data(200, "application/json", jsonBytes)
		return
	}

	response.Success(c, spec, "OpenAPI specification generated successfully")
}

// ExportPostman generates Postman collection for a project
// GET /api/v1/projects/:id/export/postman
func (h *ExportHandler) ExportPostman(c *gin.Context) {
	projectUUID := c.Param("id")
	if projectUUID == "" {
		response.BadRequest(c, "Invalid project ID", nil)
		return
	}

	collection, err := h.service.GeneratePostmanCollection(projectUUID)
	if err != nil {
		response.InternalError(c, "Failed to generate Postman collection", err)
		return
	}

	// Check if download is requested
	if c.Query("download") == "true" {
		jsonBytes, err := json.MarshalIndent(collection, "", "  ")
		if err != nil {
			response.InternalError(c, "Failed to serialize Postman collection", err)
			return
		}

		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="postman-%s.json"`, projectUUID[:8]))
		c.Header("Content-Type", "application/json")
		c.Data(200, "application/json", jsonBytes)
		return
	}

	response.Success(c, collection, "Postman collection generated successfully")
}
