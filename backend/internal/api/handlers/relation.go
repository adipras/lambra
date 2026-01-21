package handlers

import (
	"net/http"

	"github.com/yourusername/lambra/internal/service"
	"github.com/yourusername/lambra/pkg/response"

	"github.com/gin-gonic/gin"
)

type RelationHandler struct {
	relationService *service.RelationService
}

func NewRelationHandler(relationService *service.RelationService) *RelationHandler {
	return &RelationHandler{
		relationService: relationService,
	}
}

// CreateRelationRequest represents the request to create a relation
type CreateRelationRequest struct {
	SourceEntityID  string `json:"source_entity_id" binding:"required"`
	TargetEntityID  string `json:"target_entity_id" binding:"required"`
	FieldName       string `json:"field_name"`       // Optional, auto-generated if empty
	RelationType    string `json:"relation_type" binding:"required"`
	OnDelete        string `json:"on_delete"`        // Optional, defaults to RESTRICT
	OnUpdate        string `json:"on_update"`        // Optional, defaults to CASCADE
	JunctionTable   string `json:"junction_table"`   // Optional, auto-generated for manyToMany
	Description     string `json:"description"`      // Optional
}

// UpdateRelationRequest represents the request to update a relation
type UpdateRelationRequest struct {
	FieldName     string `json:"field_name"`
	RelationType  string `json:"relation_type"`
	OnDelete      string `json:"on_delete"`
	OnUpdate      string `json:"on_update"`
	JunctionTable string `json:"junction_table"`
	Description   string `json:"description"`
}

// CreateRelation handles POST /api/v1/relations
func (h *RelationHandler) CreateRelation(c *gin.Context) {
	var req CreateRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Set defaults
	if req.OnDelete == "" {
		req.OnDelete = "RESTRICT"
	}
	if req.OnUpdate == "" {
		req.OnUpdate = "CASCADE"
	}

	relation, err := h.relationService.CreateRelation(
		req.SourceEntityID,
		req.TargetEntityID,
		req.FieldName,
		req.RelationType,
		req.OnDelete,
		req.OnUpdate,
		req.JunctionTable,
		req.Description,
	)

	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, relation)
}

// GetRelation handles GET /api/v1/relations/:id
func (h *RelationHandler) GetRelation(c *gin.Context) {
	id := c.Param("id")

	relation, err := h.relationService.GetRelationByUUID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Relation not found")
		return
	}

	response.Success(c, relation)
}

// GetEntityRelations handles GET /api/v1/entities/:id/relations
func (h *RelationHandler) GetEntityRelations(c *gin.Context) {
	entityID := c.Param("id")

	relations, err := h.relationService.GetEntityRelations(entityID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"data": relations,
	})
}

// UpdateRelation handles PUT /api/v1/relations/:id
func (h *RelationHandler) UpdateRelation(c *gin.Context) {
	id := c.Param("id")

	var req UpdateRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	relation, err := h.relationService.UpdateRelation(
		id,
		req.FieldName,
		req.RelationType,
		req.OnDelete,
		req.OnUpdate,
		req.JunctionTable,
		req.Description,
	)

	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, relation)
}

// DeleteRelation handles DELETE /api/v1/relations/:id
func (h *RelationHandler) DeleteRelation(c *gin.Context) {
	id := c.Param("id")

	// TODO: Get actual user from auth context
	deletedBy := "system"

	err := h.relationService.DeleteRelation(id, deletedBy)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "Relation deleted successfully",
	})
}
