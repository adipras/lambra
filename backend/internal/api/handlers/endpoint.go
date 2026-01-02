package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/service"
	"github.com/yourusername/lambra/pkg/response"
)

type EndpointHandler struct {
	service           *service.EndpointService
	deploymentService *service.DeploymentService
}

func NewEndpointHandler(service *service.EndpointService, deploymentService *service.DeploymentService) *EndpointHandler {
	return &EndpointHandler{
		service:           service,
		deploymentService: deploymentService,
	}
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

// TestEndpoint tests an endpoint by sending a request to the deployed service
// POST /api/v1/endpoints/:id/test
func (h *EndpointHandler) TestEndpoint(c *gin.Context) {
	endpointUUID := c.Param("id")
	if endpointUUID == "" {
		response.BadRequest(c, "Invalid endpoint ID", nil)
		return
	}

	// Get endpoint details
	endpoint, err := h.service.GetEndpointByUUID(endpointUUID)
	if err != nil {
		response.NotFound(c, "Endpoint not found")
		return
	}

	// Get project to find deployment URL
	projectUUID, err := h.service.GetProjectUUIDByEndpointUUID(endpointUUID)
	if err != nil {
		response.InternalError(c, "Failed to get project", err)
		return
	}

	// Get deployment status to find service URL
	status, err := h.deploymentService.GetServiceStatus(c.Request.Context(), projectUUID)
	if err != nil {
		response.InternalError(c, "Failed to get deployment status", err)
		return
	}

	if status.Status != "running" {
		response.BadRequest(c, "Service is not running. Deploy the service first.", nil)
		return
	}

	// Parse request body
	var req models.TestEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	// Build URL - use InternalURL for backend-to-service communication
	targetURL := status.InternalURL + endpoint.Path

	// Add query parameters if provided
	// Query params are appended to the URL as ?key=value&key2=value2
	if len(req.Params) > 0 {
		queryParams := make([]string, 0, len(req.Params))
		for key, value := range req.Params {
			queryParams = append(queryParams, key+"="+value)
		}
		if strings.Contains(targetURL, "?") {
			targetURL += "&" + strings.Join(queryParams, "&")
		} else {
			targetURL += "?" + strings.Join(queryParams, "&")
		}
	}

	// Create HTTP request
	var bodyReader io.Reader
	if len(req.Body) > 0 {
		bodyReader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(endpoint.Method, targetURL, bodyReader)
	if err != nil {
		response.InternalError(c, "Failed to create request", err)
		return
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Send request and measure time
	client := &http.Client{Timeout: 30 * time.Second}
	startTime := time.Now()
	httpResp, err := client.Do(httpReq)
	responseTime := time.Since(startTime).Milliseconds()

	if err != nil {
		response.Success(c, &models.TestEndpointResponse{
			StatusCode:   0,
			ResponseTime: responseTime,
			Error:        err.Error(),
		}, "Test completed with error")
		return
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		response.InternalError(c, "Failed to read response", err)
		return
	}

	// Build response headers map
	respHeaders := make(map[string]string)
	for key := range httpResp.Header {
		respHeaders[key] = httpResp.Header.Get(key)
	}

	// Parse response body as JSON if possible
	var jsonBody json.RawMessage
	if err := json.Unmarshal(respBody, &jsonBody); err != nil {
		// If not valid JSON, wrap as string
		jsonBody = json.RawMessage(`"` + string(respBody) + `"`)
	} else {
		jsonBody = respBody
	}

	testResponse := &models.TestEndpointResponse{
		StatusCode:   httpResp.StatusCode,
		ResponseTime: responseTime,
		Headers:      respHeaders,
		Body:         jsonBody,
	}

	response.Success(c, testResponse, "Endpoint test completed")
}
