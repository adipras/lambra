package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/lambra/internal/service"
	"github.com/yourusername/lambra/pkg/response"
)

type SnapshotHandler struct {
	snapshotService   *service.SnapshotService
	deploymentService *service.DeploymentService
}

func NewSnapshotHandler(snapshotService *service.SnapshotService, deploymentService *service.DeploymentService) *SnapshotHandler {
	return &SnapshotHandler{
		snapshotService:   snapshotService,
		deploymentService: deploymentService,
	}
}

// ListByProject returns all snapshots for a project
func (h *SnapshotHandler) ListByProject(c *gin.Context) {
	projectUUID := c.Param("id")
	if projectUUID == "" {
		response.BadRequest(c, "Project ID is required", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	snapshots, total, err := h.snapshotService.GetSnapshotsByProject(projectUUID, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to retrieve snapshots", err)
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	response.SuccessWithPagination(c, snapshots, response.Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	}, "Snapshots retrieved successfully")
}

// Get returns a single snapshot by UUID
func (h *SnapshotHandler) Get(c *gin.Context) {
	snapshotUUID := c.Param("id")
	if snapshotUUID == "" {
		response.BadRequest(c, "Snapshot ID is required", nil)
		return
	}

	snapshot, err := h.snapshotService.GetSnapshot(snapshotUUID)
	if err != nil {
		response.NotFound(c, "Snapshot not found")
		return
	}

	response.Success(c, snapshot, "Snapshot retrieved successfully")
}

// Create creates a new snapshot for a project
func (h *SnapshotHandler) Create(c *gin.Context) {
	projectUUID := c.Param("id")
	if projectUUID == "" {
		response.BadRequest(c, "Project ID is required", nil)
		return
	}

	// Get created_by from request or default to "system"
	createdBy := c.DefaultQuery("created_by", "system")

	snapshot, err := h.snapshotService.CreateSnapshot(projectUUID, createdBy)
	if err != nil {
		response.InternalError(c, "Failed to create snapshot", err)
		return
	}

	response.Created(c, snapshot, "Snapshot created successfully")
}

// Rollback restores entities/endpoints from a snapshot and redeploys
func (h *SnapshotHandler) Rollback(c *gin.Context) {
	snapshotUUID := c.Param("id")
	if snapshotUUID == "" {
		response.BadRequest(c, "Snapshot ID is required", nil)
		return
	}

	rolledBackBy := c.DefaultQuery("rolled_back_by", "system")

	// Rollback to snapshot (restore entities and endpoints)
	projectUUID, err := h.snapshotService.RollbackToSnapshot(snapshotUUID, rolledBackBy)
	if err != nil {
		response.InternalError(c, "Failed to rollback to snapshot", err)
		return
	}

	// Redeploy the service with restored configuration
	result, err := h.deploymentService.RedeployService(c.Request.Context(), projectUUID)
	if err != nil {
		response.InternalError(c, "Rollback succeeded but redeployment failed", err)
		return
	}

	response.Success(c, gin.H{
		"message":      "Rollback and redeployment successful",
		"project_id":   projectUUID,
		"snapshot_id":  snapshotUUID,
		"service_url":  result.URL,
		"internal_url": result.InternalURL,
	}, "Rollback completed successfully")
}

// Delete soft deletes a snapshot
func (h *SnapshotHandler) Delete(c *gin.Context) {
	snapshotUUID := c.Param("id")
	if snapshotUUID == "" {
		response.BadRequest(c, "Snapshot ID is required", nil)
		return
	}

	deletedBy := c.DefaultQuery("deleted_by", "system")

	err := h.snapshotService.DeleteSnapshot(snapshotUUID, deletedBy)
	if err != nil {
		response.InternalError(c, "Failed to delete snapshot", err)
		return
	}

	response.Success(c, nil, "Snapshot deleted successfully")
}

// GetMetadata returns the parsed metadata from a snapshot
func (h *SnapshotHandler) GetMetadata(c *gin.Context) {
	snapshotUUID := c.Param("id")
	if snapshotUUID == "" {
		response.BadRequest(c, "Snapshot ID is required", nil)
		return
	}

	metadata, err := h.snapshotService.GetSnapshotMetadata(snapshotUUID)
	if err != nil {
		response.NotFound(c, "Snapshot not found or invalid metadata")
		return
	}

	response.Success(c, metadata, "Snapshot metadata retrieved successfully")
}
