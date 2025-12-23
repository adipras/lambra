package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
)

// ExportService handles API documentation export
type ExportService struct {
	projectRepo  *repository.ProjectRepository
	entityRepo   *repository.EntityRepository
	endpointRepo *repository.EndpointRepository
}

// NewExportService creates a new export service
func NewExportService(
	projectRepo *repository.ProjectRepository,
	entityRepo *repository.EntityRepository,
	endpointRepo *repository.EndpointRepository,
) *ExportService {
	return &ExportService{
		projectRepo:  projectRepo,
		entityRepo:   entityRepo,
		endpointRepo: endpointRepo,
	}
}

// OpenAPISpec represents an OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       OpenAPIInfo            `json:"info"`
	Servers    []OpenAPIServer        `json:"servers,omitempty"`
	Paths      map[string]interface{} `json:"paths"`
	Components OpenAPIComponents      `json:"components,omitempty"`
}

type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type OpenAPIComponents struct {
	Schemas         map[string]interface{} `json:"schemas,omitempty"`
	SecuritySchemes map[string]interface{} `json:"securitySchemes,omitempty"`
}

// PostmanCollection represents a Postman collection
type PostmanCollection struct {
	Info     PostmanInfo      `json:"info"`
	Item     []PostmanItem    `json:"item"`
	Variable []PostmanVar     `json:"variable,omitempty"`
	Auth     *PostmanAuth     `json:"auth,omitempty"`
}

type PostmanInfo struct {
	PostmanID   string `json:"_postman_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Schema      string `json:"schema"`
}

type PostmanItem struct {
	Name        string             `json:"name"`
	Item        []PostmanItem      `json:"item,omitempty"`
	Request     *PostmanRequest    `json:"request,omitempty"`
	Response    []PostmanResponse  `json:"response,omitempty"`
	Description string             `json:"description,omitempty"`
}

type PostmanRequest struct {
	Method string             `json:"method"`
	Header []PostmanHeader    `json:"header,omitempty"`
	Body   *PostmanBody       `json:"body,omitempty"`
	URL    PostmanURL         `json:"url"`
	Auth   *PostmanAuth       `json:"auth,omitempty"`
}

type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

type PostmanBody struct {
	Mode string `json:"mode"`
	Raw  string `json:"raw,omitempty"`
}

type PostmanURL struct {
	Raw   string   `json:"raw"`
	Host  []string `json:"host,omitempty"`
	Path  []string `json:"path,omitempty"`
	Query []PostmanQuery `json:"query,omitempty"`
}

type PostmanQuery struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PostmanAuth struct {
	Type   string        `json:"type"`
	Bearer []PostmanVar  `json:"bearer,omitempty"`
}

type PostmanVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

type PostmanResponse struct{}

// GenerateOpenAPISpec generates OpenAPI 3.0 specification for a project
func (s *ExportService) GenerateOpenAPISpec(projectUUID string) (*OpenAPISpec, error) {
	// Get project
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Get entities
	entities, err := s.entityRepo.GetByProjectID(project.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	// Build paths and schemas
	paths := make(map[string]interface{})
	schemas := make(map[string]interface{})

	for _, entity := range entities {
		// Get endpoints for this entity
		endpoints, err := s.endpointRepo.GetByEntityID(entity.ID)
		if err != nil {
			continue
		}

		// Generate schema for entity
		entitySchema := s.generateEntitySchema(&entity)
		schemas[entity.Name] = entitySchema
		schemas[entity.Name+"Request"] = s.generateRequestSchema(&entity)
		schemas[entity.Name+"Response"] = s.generateResponseSchema(&entity)

		// Generate paths for endpoints
		for _, endpoint := range endpoints {
			s.addEndpointToPath(paths, &endpoint, &entity)
		}
	}

	spec := &OpenAPISpec{
		OpenAPI: "3.0.3",
		Info: OpenAPIInfo{
			Title:       project.Name + " API",
			Description: project.Description.String,
			Version:     "1.0.0",
		},
		Servers: []OpenAPIServer{
			{
				URL:         "http://localhost:8080",
				Description: "Development server",
			},
		},
		Paths: paths,
		Components: OpenAPIComponents{
			Schemas: schemas,
			SecuritySchemes: map[string]interface{}{
				"bearerAuth": map[string]interface{}{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
		},
	}

	return spec, nil
}

// GeneratePostmanCollection generates a Postman collection for a project
func (s *ExportService) GeneratePostmanCollection(projectUUID string) (*PostmanCollection, error) {
	// Get project
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Get entities
	entities, err := s.entityRepo.GetByProjectID(project.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	// Build items
	items := []PostmanItem{}

	for _, entity := range entities {
		// Get endpoints for this entity
		endpoints, err := s.endpointRepo.GetByEntityID(entity.ID)
		if err != nil {
			continue
		}

		// Create folder for entity
		entityFolder := PostmanItem{
			Name:        entity.Name,
			Description: entity.Description.String,
			Item:        []PostmanItem{},
		}

		// Add endpoints to folder
		for _, endpoint := range endpoints {
			item := s.createPostmanItem(&endpoint, &entity)
			entityFolder.Item = append(entityFolder.Item, item)
		}

		if len(entityFolder.Item) > 0 {
			items = append(items, entityFolder)
		}
	}

	collection := &PostmanCollection{
		Info: PostmanInfo{
			PostmanID:   project.UUID,
			Name:        project.Name + " API",
			Description: project.Description.String,
			Schema:      "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		Item: items,
		Variable: []PostmanVar{
			{
				Key:   "baseUrl",
				Value: "http://localhost:8080",
				Type:  "string",
			},
		},
		Auth: &PostmanAuth{
			Type: "bearer",
			Bearer: []PostmanVar{
				{
					Key:   "token",
					Value: "{{authToken}}",
					Type:  "string",
				},
			},
		},
	}

	return collection, nil
}

// Helper functions

func (s *ExportService) generateEntitySchema(entity *models.Entity) map[string]interface{} {
	properties := make(map[string]interface{})
	required := []string{}

	// Add id field
	properties["id"] = map[string]interface{}{
		"type":        "string",
		"format":      "uuid",
		"description": "Unique identifier",
	}

	// Parse entity fields from JSON
	var fields []models.EntityField
	if len(entity.Fields) > 0 {
		json.Unmarshal(entity.Fields, &fields)
	}

	// Add entity fields
	for _, field := range fields {
		prop := map[string]interface{}{
			"type": s.jsonSchemaType(field.Type),
		}
		if field.Description != "" {
			prop["description"] = field.Description
		}
		if field.Length > 0 {
			prop["maxLength"] = field.Length
		}
		properties[field.Name] = prop

		if field.Required {
			required = append(required, field.Name)
		}
	}

	// Add timestamp fields
	properties["created_at"] = map[string]interface{}{
		"type":   "string",
		"format": "date-time",
	}
	properties["updated_at"] = map[string]interface{}{
		"type":   "string",
		"format": "date-time",
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

func (s *ExportService) generateRequestSchema(entity *models.Entity) map[string]interface{} {
	properties := make(map[string]interface{})
	required := []string{}

	// Parse entity fields from JSON
	var fields []models.EntityField
	if len(entity.Fields) > 0 {
		json.Unmarshal(entity.Fields, &fields)
	}

	for _, field := range fields {
		prop := map[string]interface{}{
			"type": s.jsonSchemaType(field.Type),
		}
		if field.Description != "" {
			prop["description"] = field.Description
		}
		if field.Length > 0 {
			prop["maxLength"] = field.Length
		}
		properties[field.Name] = prop

		if field.Required {
			required = append(required, field.Name)
		}
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

func (s *ExportService) generateResponseSchema(entity *models.Entity) map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type": "boolean",
			},
			"message": map[string]interface{}{
				"type": "string",
			},
			"data": map[string]interface{}{
				"$ref": "#/components/schemas/" + entity.Name,
			},
		},
	}
}

func (s *ExportService) addEndpointToPath(paths map[string]interface{}, endpoint *models.Endpoint, entity *models.Entity) {
	path := endpoint.Path
	method := strings.ToLower(endpoint.Method)

	// Initialize path if not exists
	if _, exists := paths[path]; !exists {
		paths[path] = make(map[string]interface{})
	}

	// Build operation
	operation := map[string]interface{}{
		"summary":     endpoint.Name,
		"description": endpoint.Description.String,
		"tags":        []string{entity.Name},
		"operationId": toCamelCase(endpoint.Name),
	}

	// Add request body for POST/PUT/PATCH
	if method == "post" || method == "put" || method == "patch" {
		var requestSchema interface{}
		if len(endpoint.RequestSchema) > 0 {
			json.Unmarshal(endpoint.RequestSchema, &requestSchema)
		} else {
			requestSchema = map[string]interface{}{
				"$ref": "#/components/schemas/" + entity.Name + "Request",
			}
		}
		operation["requestBody"] = map[string]interface{}{
			"required": true,
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": requestSchema,
				},
			},
		}
	}

	// Add responses
	var responseSchema interface{}
	if len(endpoint.ResponseSchema) > 0 {
		json.Unmarshal(endpoint.ResponseSchema, &responseSchema)
	} else {
		responseSchema = map[string]interface{}{
			"$ref": "#/components/schemas/" + entity.Name + "Response",
		}
	}

	operation["responses"] = map[string]interface{}{
		"200": map[string]interface{}{
			"description": "Successful response",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": responseSchema,
				},
			},
		},
		"400": map[string]interface{}{
			"description": "Bad request",
		},
		"404": map[string]interface{}{
			"description": "Not found",
		},
		"500": map[string]interface{}{
			"description": "Internal server error",
		},
	}

	// Add security if required
	if endpoint.RequireAuth {
		operation["security"] = []map[string]interface{}{
			{"bearerAuth": []string{}},
		}
	}

	paths[path].(map[string]interface{})[method] = operation
}

func (s *ExportService) createPostmanItem(endpoint *models.Endpoint, entity *models.Entity) PostmanItem {
	// Build URL path
	path := endpoint.Path
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	pathParts := strings.Split(path, "/")

	item := PostmanItem{
		Name:        endpoint.Name,
		Description: endpoint.Description.String,
		Request: &PostmanRequest{
			Method: endpoint.Method,
			Header: []PostmanHeader{
				{
					Key:   "Content-Type",
					Value: "application/json",
					Type:  "text",
				},
			},
			URL: PostmanURL{
				Raw:  "{{baseUrl}}/" + path,
				Host: []string{"{{baseUrl}}"},
				Path: pathParts,
			},
		},
	}

	// Add request body for POST/PUT/PATCH
	if endpoint.Method == "POST" || endpoint.Method == "PUT" || endpoint.Method == "PATCH" {
		var bodyContent string
		if len(endpoint.RequestSchema) > 0 {
			// Use request schema as example body
			var schema map[string]interface{}
			if err := json.Unmarshal(endpoint.RequestSchema, &schema); err == nil {
				example := s.generateExampleFromSchema(schema)
				if exampleBytes, err := json.MarshalIndent(example, "", "  "); err == nil {
					bodyContent = string(exampleBytes)
				}
			}
		}
		if bodyContent == "" {
			bodyContent = "{}"
		}
		item.Request.Body = &PostmanBody{
			Mode: "raw",
			Raw:  bodyContent,
		}
	}

	// Add auth if required
	if endpoint.RequireAuth {
		item.Request.Auth = &PostmanAuth{
			Type: "bearer",
			Bearer: []PostmanVar{
				{
					Key:   "token",
					Value: "{{authToken}}",
					Type:  "string",
				},
			},
		}
	}

	return item
}

func (s *ExportService) generateExampleFromSchema(schema map[string]interface{}) map[string]interface{} {
	example := make(map[string]interface{})

	if props, ok := schema["properties"].(map[string]interface{}); ok {
		for key, val := range props {
			if propMap, ok := val.(map[string]interface{}); ok {
				example[key] = s.getExampleValue(propMap)
			}
		}
	} else {
		// Flat schema (key -> type)
		for key, val := range schema {
			if typeStr, ok := val.(string); ok {
				example[key] = s.getExampleValueForType(typeStr)
			}
		}
	}

	return example
}

func (s *ExportService) getExampleValue(prop map[string]interface{}) interface{} {
	propType, _ := prop["type"].(string)
	return s.getExampleValueForType(propType)
}

func (s *ExportService) getExampleValueForType(propType string) interface{} {
	switch propType {
	case "string":
		return "example"
	case "number", "integer", "int", "float":
		return 0
	case "boolean", "bool":
		return false
	case "array":
		return []interface{}{}
	case "object":
		return map[string]interface{}{}
	default:
		return nil
	}
}

func (s *ExportService) jsonSchemaType(fieldType string) string {
	switch strings.ToLower(fieldType) {
	case "string", "text", "varchar", "char":
		return "string"
	case "int", "integer", "bigint", "smallint", "tinyint":
		return "integer"
	case "float", "double", "decimal", "number":
		return "number"
	case "bool", "boolean":
		return "boolean"
	case "date":
		return "string" // with format: date
	case "datetime", "timestamp":
		return "string" // with format: date-time
	case "json", "jsonb":
		return "object"
	case "array":
		return "array"
	default:
		return "string"
	}
}

func toCamelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '_' || r == '-'
	})

	for i := range words {
		if i == 0 {
			words[i] = strings.ToLower(words[i])
		} else {
			words[i] = strings.Title(words[i])
		}
	}

	return strings.Join(words, "")
}
