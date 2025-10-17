package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourusername/lambra/internal/models"
)

// CodeGenerator generates code from entities
type CodeGenerator struct {
	engine    *TemplateEngine
	templates map[string]string
}

// NewCodeGenerator creates a new code generator
func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		engine:    NewTemplateEngine(),
		templates: make(map[string]string),
	}
}

// GenerateContext holds all data needed for code generation
type GenerateContext struct {
	Project      *models.Project
	Entity       *models.Entity
	Fields       []FieldContext
	PackageName  string
	Imports      []string
	EntityName   string
	EntityNameLC string
	TableName    string
	HasUUID      bool
	HasTimestamp bool
}

// FieldContext represents a field for template rendering
type FieldContext struct {
	Name         string
	NameLC       string
	Type         string
	GoType       string
	JSONTag      string
	DBTag        string
	ValidateTag  string
	Required     bool
	Nullable     bool
	DefaultValue string
	Description  string
}

// GenerateAll generates all code for an entity
func (g *CodeGenerator) GenerateAll(ctx *GenerateContext, outputDir string) error {
	generators := []struct {
		name     string
		template string
		filename string
	}{
		{"model", modelTemplate, fmt.Sprintf("%s.go", toSnakeCase(ctx.EntityName))},
		{"repository", repositoryTemplate, fmt.Sprintf("%s_repository.go", toSnakeCase(ctx.EntityName))},
		{"service", serviceTemplate, fmt.Sprintf("%s_service.go", toSnakeCase(ctx.EntityName))},
		{"handler", handlerTemplate, fmt.Sprintf("%s_handler.go", toSnakeCase(ctx.EntityName))},
		{"dto", dtoTemplate, fmt.Sprintf("%s_dto.go", toSnakeCase(ctx.EntityName))},
	}

	for _, gen := range generators {
		code, err := g.engine.Render(gen.template, ctx)
		if err != nil {
			return fmt.Errorf("failed to generate %s: %w", gen.name, err)
		}

		// Determine output path based on layer
		var layerDir string
		switch gen.name {
		case "model":
			layerDir = "models"
		case "repository":
			layerDir = "repository"
		case "service":
			layerDir = "service"
		case "handler":
			layerDir = filepath.Join("api", "handlers")
		case "dto":
			layerDir = filepath.Join("api", "dto")
		}

		outputPath := filepath.Join(outputDir, layerDir, gen.filename)
		if err := g.writeFile(outputPath, code); err != nil {
			return fmt.Errorf("failed to write %s: %w", gen.name, err)
		}
	}

	return nil
}

// GenerateModel generates model code
func (g *CodeGenerator) GenerateModel(ctx *GenerateContext) (string, error) {
	return g.engine.Render(modelTemplate, ctx)
}

// GenerateRepository generates repository code
func (g *CodeGenerator) GenerateRepository(ctx *GenerateContext) (string, error) {
	return g.engine.Render(repositoryTemplate, ctx)
}

// GenerateService generates service code
func (g *CodeGenerator) GenerateService(ctx *GenerateContext) (string, error) {
	return g.engine.Render(serviceTemplate, ctx)
}

// GenerateHandler generates handler code
func (g *CodeGenerator) GenerateHandler(ctx *GenerateContext) (string, error) {
	return g.engine.Render(handlerTemplate, ctx)
}

// GenerateDTO generates DTO code
func (g *CodeGenerator) GenerateDTO(ctx *GenerateContext) (string, error) {
	return g.engine.Render(dtoTemplate, ctx)
}

// GenerateMigration generates database migration
func (g *CodeGenerator) GenerateMigration(ctx *GenerateContext) (up string, down string, err error) {
	up, err = g.engine.Render(migrationUpTemplate, ctx)
	if err != nil {
		return "", "", err
	}

	down, err = g.engine.Render(migrationDownTemplate, ctx)
	if err != nil {
		return "", "", err
	}

	return up, down, nil
}

// PrepareContext prepares generation context from entity
func (g *CodeGenerator) PrepareContext(project *models.Project, entity *models.Entity) (*GenerateContext, error) {
	ctx := &GenerateContext{
		Project:      project,
		Entity:       entity,
		EntityName:   entity.Name,
		EntityNameLC: toCamelCase(entity.Name),
		TableName:    entity.TableName,
		PackageName:  "models",
		HasUUID:      true,
		HasTimestamp: true,
	}

	// Parse fields from JSON
	var fields []models.EntityField
	if err := json.Unmarshal(entity.Fields, &fields); err != nil {
		return nil, fmt.Errorf("failed to parse entity fields: %w", err)
	}

	for _, f := range fields {
		field := g.parseField(f)
		ctx.Fields = append(ctx.Fields, field)
	}

	// Determine imports
	ctx.Imports = g.determineImports(ctx)

	return ctx, nil
}

// parseField parses a field into FieldContext
func (g *CodeGenerator) parseField(field models.EntityField) FieldContext {
	goType := toGoType(field.Type)
	nullable := !field.Required
	if nullable {
		goType = "*" + goType
	}

	validateRules := ""
	if field.Required {
		validateRules = "required"
	}
	if field.Length > 0 {
		if validateRules != "" {
			validateRules += ","
		}
		validateRules += fmt.Sprintf("max=%d", field.Length)
	}

	validateTag := ""
	if validateRules != "" {
		validateTag = fmt.Sprintf(`validate:"%s"`, validateRules)
	}

	return FieldContext{
		Name:         toPascalCase(field.Name),
		NameLC:       toCamelCase(field.Name),
		Type:         field.Type,
		GoType:       goType,
		JSONTag:      toJSONTag(field.Name, field.Required),
		DBTag:        toDBTag(field.Name),
		ValidateTag:  validateTag,
		Required:     field.Required,
		Nullable:     nullable,
		DefaultValue: field.DefaultValue,
		Description:  field.Description,
	}
}

// determineImports determines required imports based on context
func (g *CodeGenerator) determineImports(ctx *GenerateContext) []string {
	imports := make(map[string]bool)

	// Always needed
	imports["time"] = true

	// Check if UUID is used
	if ctx.HasUUID {
		imports["github.com/google/uuid"] = true
	}

	// Check field types
	for _, field := range ctx.Fields {
		switch field.Type {
		case "json":
			imports["encoding/json"] = true
		case "uuid":
			imports["github.com/google/uuid"] = true
		case "date", "datetime", "timestamp":
			imports["time"] = true
		}
	}

	// Convert to slice
	result := make([]string, 0, len(imports))
	for imp := range imports {
		result = append(result, imp)
	}

	return result
}

// writeFile writes code to a file
func (g *CodeGenerator) writeFile(path string, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// FormatCode formats generated code using gofmt
func (g *CodeGenerator) FormatCode(code string) (string, error) {
	// This would use golang.org/x/tools/imports or gofmt
	// For now, return as is
	return code, nil
}

// ValidateContext validates the generation context
func (g *CodeGenerator) ValidateContext(ctx *GenerateContext) error {
	if ctx.EntityName == "" {
		return fmt.Errorf("entity name is required")
	}
	if ctx.TableName == "" {
		return fmt.Errorf("table name is required")
	}
	if len(ctx.Fields) == 0 {
		return fmt.Errorf("at least one field is required")
	}
	return nil
}

// GetGeneratedFiles returns list of files that will be generated
func (g *CodeGenerator) GetGeneratedFiles(entityName string) []string {
	snake := toSnakeCase(entityName)
	return []string{
		fmt.Sprintf("models/%s.go", snake),
		fmt.Sprintf("repository/%s_repository.go", snake),
		fmt.Sprintf("service/%s_service.go", snake),
		fmt.Sprintf("api/handlers/%s_handler.go", snake),
		fmt.Sprintf("api/dto/%s_dto.go", snake),
	}
}
