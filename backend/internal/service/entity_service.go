package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
)

type EntityService struct {
	repo         *repository.EntityRepository
	projectRepo  *repository.ProjectRepository
	endpointRepo *repository.EndpointRepository
}

func NewEntityService(repo *repository.EntityRepository, projectRepo *repository.ProjectRepository, endpointRepo *repository.EndpointRepository) *EntityService {
	return &EntityService{
		repo:         repo,
		projectRepo:  projectRepo,
		endpointRepo: endpointRepo,
	}
}

func (s *EntityService) CreateEntity(req *models.CreateEntityRequest) (*models.Entity, error) {
	// Validate project exists and get internal ID
	project, err := s.projectRepo.GetByUUID(req.ProjectUUID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Marshal fields to JSON
	fieldsJSON, err := json.Marshal(req.Fields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields: %w", err)
	}

	entity := &models.Entity{
		ProjectID: project.ID, // Use internal project ID
		Name:      req.Name,
		TableName: req.TableName,
		Fields:    fieldsJSON,
	}

	if req.Description != "" {
		entity.Description.String = req.Description
		entity.Description.Valid = true
	}

	// Set created_by (in future, get from auth context)
	entity.SetCreatedBy("system")

	err = s.repo.Create(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	// Generate CRUD endpoints if requested
	if req.GenerateEndpoints != nil {
		if err := s.generateCRUDEndpoints(entity, project.ID, req.GenerateEndpoints); err != nil {
			// Log error but don't fail entity creation
			fmt.Printf("Warning: failed to generate endpoints: %v\n", err)
		}
	}

	return entity, nil
}

// generateCRUDEndpoints creates CRUD endpoints for an entity based on options
func (s *EntityService) generateCRUDEndpoints(entity *models.Entity, projectID int64, opts *models.GenerateEndpoints) error {
	entityName := toPascalCase(entity.Name)
	entityNamePlural := pluralize(entityName)
	entityNameSnake := toSnakeCase(entity.Name)
	entityNameLower := strings.ToLower(entityName)
	basePath := "/" + pluralize(entityNameSnake)

	// Parse entity fields from JSON
	var fields []models.EntityField
	if err := json.Unmarshal(entity.Fields, &fields); err != nil {
		fields = []models.EntityField{} // fallback to empty if parse fails
	}

	// Generate schemas based on entity fields
	requestSchema := s.generateRequestSchema(fields)
	responseSchema := s.generateResponseSchema(fields, entityNameLower)
	listResponseSchema := s.generateListResponseSchema(fields, entityNameLower)

	// Generate query param schema for id parameter
	idQueryParamSchema := s.generateIDQueryParamSchema()

	endpoints := []struct {
		enabled        bool
		name           string
		method         string
		path           string
		description    string
		requestSchema  json.RawMessage
		responseSchema json.RawMessage
	}{
		{opts.List, "List" + entityNamePlural, "GET", basePath, fmt.Sprintf("Get all %s", strings.ToLower(entityNamePlural)), nil, listResponseSchema},
		{opts.Get, "Get" + entityName, "GET", basePath + "/detail", fmt.Sprintf("Get %s by ID (query param: id)", entityNameLower), idQueryParamSchema, responseSchema},
		{opts.Create, "Create" + entityName, "POST", basePath, fmt.Sprintf("Create a new %s", entityNameLower), requestSchema, responseSchema},
		{opts.Update, "Update" + entityName, "PUT", basePath + "/update", fmt.Sprintf("Update %s by ID (query param: id)", entityNameLower), s.mergeSchemas(idQueryParamSchema, requestSchema), responseSchema},
		{opts.Delete, "Delete" + entityName, "DELETE", basePath + "/delete", fmt.Sprintf("Delete %s by ID (query param: id)", entityNameLower), idQueryParamSchema, s.generateDeleteResponseSchema()},
	}

	for _, ep := range endpoints {
		if !ep.enabled {
			continue
		}

		endpoint := &models.Endpoint{
			EntityID:       entity.ID,
			ProjectID:      projectID,
			Name:           ep.name,
			Method:         ep.method,
			Path:           ep.path,
			RequireAuth:    false,
			RequestSchema:  ep.requestSchema,
			ResponseSchema: ep.responseSchema,
		}
		endpoint.Description.String = ep.description
		endpoint.Description.Valid = true
		endpoint.SetCreatedBy("system")

		if err := s.endpointRepo.Create(endpoint); err != nil {
			return fmt.Errorf("failed to create endpoint %s: %w", ep.name, err)
		}
	}

	return nil
}

// generateRequestSchema creates JSON schema for create/update request body
func (s *EntityService) generateRequestSchema(fields []models.EntityField) json.RawMessage {
	properties := make(map[string]interface{})
	required := []string{}

	for _, field := range fields {
		fieldSchema := map[string]interface{}{
			"type": fieldTypeToJSONType(field.Type),
		}
		if field.Description != "" {
			fieldSchema["description"] = field.Description
		}
		if field.Length > 0 && field.Type == "string" {
			fieldSchema["maxLength"] = field.Length
		}

		// Add example values
		fieldSchema["example"] = getExampleValue(field)

		properties[toSnakeCase(field.Name)] = fieldSchema

		if field.Required {
			required = append(required, toSnakeCase(field.Name))
		}
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}

	data, _ := json.Marshal(schema)
	return data
}

// generateResponseSchema creates JSON schema for single entity response
func (s *EntityService) generateResponseSchema(fields []models.EntityField, entityName string) json.RawMessage {
	properties := map[string]interface{}{
		"id":   map[string]interface{}{"type": "integer", "description": "Unique identifier", "example": 1},
		"uuid": map[string]interface{}{"type": "string", "format": "uuid", "description": "UUID identifier", "example": "550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, field := range fields {
		fieldSchema := map[string]interface{}{
			"type": fieldTypeToJSONType(field.Type),
		}
		if field.Description != "" {
			fieldSchema["description"] = field.Description
		}
		fieldSchema["example"] = getExampleValue(field)
		properties[toSnakeCase(field.Name)] = fieldSchema
	}

	properties["created_at"] = map[string]interface{}{"type": "string", "format": "date-time", "example": "2025-01-01T00:00:00Z"}
	properties["updated_at"] = map[string]interface{}{"type": "string", "format": "date-time", "example": "2025-01-01T00:00:00Z"}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	data, _ := json.Marshal(schema)
	return data
}

// generateListResponseSchema creates JSON schema for list response with pagination
func (s *EntityService) generateListResponseSchema(fields []models.EntityField, entityName string) json.RawMessage {
	itemProperties := map[string]interface{}{
		"id":   map[string]interface{}{"type": "integer", "example": 1},
		"uuid": map[string]interface{}{"type": "string", "format": "uuid", "example": "550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, field := range fields {
		fieldSchema := map[string]interface{}{
			"type": fieldTypeToJSONType(field.Type),
		}
		fieldSchema["example"] = getExampleValue(field)
		itemProperties[toSnakeCase(field.Name)] = fieldSchema
	}

	itemProperties["created_at"] = map[string]interface{}{"type": "string", "format": "date-time"}
	itemProperties["updated_at"] = map[string]interface{}{"type": "string", "format": "date-time"}

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"data": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type":       "object",
					"properties": itemProperties,
				},
			},
			"total":  map[string]interface{}{"type": "integer", "description": "Total number of items", "example": 100},
			"limit":  map[string]interface{}{"type": "integer", "description": "Items per page", "example": 10},
			"offset": map[string]interface{}{"type": "integer", "description": "Offset from start", "example": 0},
		},
	}

	data, _ := json.Marshal(schema)
	return data
}

// generateDeleteResponseSchema creates JSON schema for delete response
func (s *EntityService) generateDeleteResponseSchema() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{"type": "string", "example": "Successfully deleted"},
		},
	}

	data, _ := json.Marshal(schema)
	return data
}

// generateIDQueryParamSchema creates JSON schema for id query parameter
func (s *EntityService) generateIDQueryParamSchema() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the resource (query parameter)",
				"example":     "550e8400-e29b-41d4-a716-446655440000",
			},
		},
		"required":          []string{"id"},
		"x-parameter-style": "query",
	}

	data, _ := json.Marshal(schema)
	return data
}

// mergeSchemas merges two JSON schemas into one
func (s *EntityService) mergeSchemas(schema1, schema2 json.RawMessage) json.RawMessage {
	if schema1 == nil {
		return schema2
	}
	if schema2 == nil {
		return schema1
	}

	var s1, s2 map[string]interface{}
	json.Unmarshal(schema1, &s1)
	json.Unmarshal(schema2, &s2)

	// Get properties from both schemas
	props1, _ := s1["properties"].(map[string]interface{})
	props2, _ := s2["properties"].(map[string]interface{})

	// Merge properties
	mergedProps := make(map[string]interface{})
	for k, v := range props1 {
		mergedProps[k] = v
	}
	for k, v := range props2 {
		mergedProps[k] = v
	}

	// Merge required fields
	var required []string
	if req1, ok := s1["required"].([]interface{}); ok {
		for _, r := range req1 {
			if str, ok := r.(string); ok {
				required = append(required, str)
			}
		}
	}
	if req2, ok := s2["required"].([]interface{}); ok {
		for _, r := range req2 {
			if str, ok := r.(string); ok {
				required = append(required, str)
			}
		}
	}

	merged := map[string]interface{}{
		"type":       "object",
		"properties": mergedProps,
	}
	if len(required) > 0 {
		merged["required"] = required
	}
	// Mark as having query params
	if s1["x-parameter-style"] == "query" || s2["x-parameter-style"] == "query" {
		merged["x-parameter-style"] = "query"
	}

	data, _ := json.Marshal(merged)
	return data
}

// fieldTypeToJSONType converts entity field type to JSON Schema type
func fieldTypeToJSONType(fieldType string) string {
	switch strings.ToLower(fieldType) {
	case "string", "text", "uuid":
		return "string"
	case "int", "integer", "bigint":
		return "integer"
	case "float", "decimal", "double":
		return "number"
	case "bool", "boolean":
		return "boolean"
	case "date", "datetime", "timestamp":
		return "string" // with format: date-time
	case "json":
		return "object"
	default:
		return "string"
	}
}

// getExampleValue returns example value based on field type and name
func getExampleValue(field models.EntityField) interface{} {
	name := strings.ToLower(field.Name)

	// Smart example based on field name
	switch {
	case strings.Contains(name, "email"):
		return "user@example.com"
	case strings.Contains(name, "name"):
		return "Example Name"
	case strings.Contains(name, "phone"):
		return "+1234567890"
	case strings.Contains(name, "price"):
		return 99.99
	case strings.Contains(name, "quantity") || strings.Contains(name, "stock") || strings.Contains(name, "count"):
		return 100
	case strings.Contains(name, "url") || strings.Contains(name, "link"):
		return "https://example.com"
	case strings.Contains(name, "description"):
		return "A detailed description"
	case strings.Contains(name, "sku") || strings.Contains(name, "code"):
		return "SKU-001"
	case strings.Contains(name, "status"):
		return "active"
	}

	// Default based on type
	switch strings.ToLower(field.Type) {
	case "string", "text":
		return "example string"
	case "int", "integer", "bigint":
		return 1
	case "float", "decimal", "double":
		return 10.5
	case "bool", "boolean":
		return true
	case "date":
		return "2025-01-01"
	case "datetime", "timestamp":
		return "2025-01-01T00:00:00Z"
	case "json":
		return map[string]interface{}{}
	default:
		return "example"
	}
}

func (s *EntityService) GetEntityByUUID(uuid string) (*models.Entity, error) {
	return s.repo.GetByUUID(uuid)
}

func (s *EntityService) GetEntitiesByProjectUUID(projectUUID string) ([]models.Entity, error) {
	// Validate project exists and get internal ID
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return s.repo.GetByProjectID(project.ID)
}

func (s *EntityService) UpdateEntity(uuid string, req *models.UpdateEntityRequest) (*models.Entity, error) {
	entity, err := s.repo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		entity.Name = req.Name
	}
	if req.TableName != "" {
		entity.TableName = req.TableName
	}
	if req.Description != "" {
		entity.Description.String = req.Description
		entity.Description.Valid = true
	}
	if len(req.Fields) > 0 {
		fieldsJSON, err := json.Marshal(req.Fields)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal fields: %w", err)
		}
		entity.Fields = fieldsJSON
	}

	// Set updated_by (in future, get from auth context)
	entity.SetUpdatedBy("system")

	err = s.repo.Update(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	return entity, nil
}

func (s *EntityService) DeleteEntity(uuid string) error {
	_, err := s.repo.GetByUUID(uuid)
	if err != nil {
		return err
	}

	// Soft delete with deleted_by (in future, get from auth context)
	return s.repo.DeleteByUUID(uuid, "system")
}
