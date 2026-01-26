package service

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/yourusername/lambra/internal/generator"
	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
)

// GeneratorService handles code generation operations
type GeneratorService struct {
	projectRepo  *repository.ProjectRepository
	entityRepo   *repository.EntityRepository
	endpointRepo *repository.EndpointRepository
	relationRepo *repository.RelationRepository
	generator    *generator.CodeGenerator
}

// NewGeneratorService creates a new generator service
func NewGeneratorService(
	projectRepo *repository.ProjectRepository,
	entityRepo *repository.EntityRepository,
	endpointRepo *repository.EndpointRepository,
	relationRepo *repository.RelationRepository,
) *GeneratorService {
	return &GeneratorService{
		projectRepo:  projectRepo,
		entityRepo:   entityRepo,
		endpointRepo: endpointRepo,
		relationRepo: relationRepo,
		generator:    generator.NewCodeGenerator(),
	}
}

// GenerateCodeRequest represents a request to generate code
type GenerateCodeRequest struct {
	EntityID  int64    `json:"entity_id"`
	OutputDir string   `json:"output_dir"`
	Layers    []string `json:"layers"` // model, repository, service, handler, dto, migration
}

// GenerateCodeResponse represents the response from code generation
type GenerateCodeResponse struct {
	Files    []GeneratedFile `json:"files"`
	EntityID int64           `json:"entity_id"`
	Success  bool            `json:"success"`
	Message  string          `json:"message"`
}

// GeneratedFile represents a generated code file
type GeneratedFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Layer   string `json:"layer"`
}

// GenerateEntity generates code for a specific entity
func (s *GeneratorService) GenerateEntity(ctx context.Context, entityID int64, outputDir string) (*GenerateCodeResponse, error) {
	// Get entity
	entity, err := s.entityRepo.GetByID(entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	// Get project
	project, err := s.projectRepo.GetByID(entity.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Prepare generation context
	genCtx, err := s.generator.PrepareContext(project, entity)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare context: %w", err)
	}

	// Fetch and add relations to context
	relations, err := s.fetchRelationsForEntity(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relations: %w", err)
	}
	genCtx.Relations = relations

	// Validate context
	if err := s.generator.ValidateContext(genCtx); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Generate code
	var files []GeneratedFile

	// Generate model
	if code, err := s.generator.GenerateModel(genCtx); err == nil {
		files = append(files, GeneratedFile{
			Path:    filepath.Join(outputDir, "models", fmt.Sprintf("%s.go", generator.ToSnakeCase(entity.Name))),
			Content: code,
			Layer:   "model",
		})
	}

	// Generate repository
	if code, err := s.generator.GenerateRepository(genCtx); err == nil {
		files = append(files, GeneratedFile{
			Path:    filepath.Join(outputDir, "repository", fmt.Sprintf("%s_repository.go", generator.ToSnakeCase(entity.Name))),
			Content: code,
			Layer:   "repository",
		})
	}

	// Generate service
	if code, err := s.generator.GenerateService(genCtx); err == nil {
		files = append(files, GeneratedFile{
			Path:    filepath.Join(outputDir, "service", fmt.Sprintf("%s_service.go", generator.ToSnakeCase(entity.Name))),
			Content: code,
			Layer:   "service",
		})
	}

	// Generate handler
	if code, err := s.generator.GenerateHandler(genCtx); err == nil {
		files = append(files, GeneratedFile{
			Path:    filepath.Join(outputDir, "api/handlers", fmt.Sprintf("%s_handler.go", generator.ToSnakeCase(entity.Name))),
			Content: code,
			Layer:   "handler",
		})
	}

	// Generate DTO
	if code, err := s.generator.GenerateDTO(genCtx); err == nil {
		files = append(files, GeneratedFile{
			Path:    filepath.Join(outputDir, "api/dto", fmt.Sprintf("%s_dto.go", generator.ToSnakeCase(entity.Name))),
			Content: code,
			Layer:   "dto",
		})
	}

	// Generate migration
	up, down, err := s.generator.GenerateMigration(genCtx)
	if err == nil {
		files = append(files, GeneratedFile{
			Path:    filepath.Join(outputDir, "migrations", fmt.Sprintf("create_%s.up.sql", generator.ToSnakeCase(entity.TableName))),
			Content: up,
			Layer:   "migration",
		})
		files = append(files, GeneratedFile{
			Path:    filepath.Join(outputDir, "migrations", fmt.Sprintf("create_%s.down.sql", generator.ToSnakeCase(entity.TableName))),
			Content: down,
			Layer:   "migration",
		})
	}

	return &GenerateCodeResponse{
		Files:    files,
		EntityID: entityID,
		Success:  true,
		Message:  fmt.Sprintf("Successfully generated %d files for entity %s", len(files), entity.Name),
	}, nil
}

// GenerateProject generates code for all entities in a project
func (s *GeneratorService) GenerateProject(ctx context.Context, projectID int64, outputDir string) (*GenerateCodeResponse, error) {
	// Get project
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Get all entities for the project
	entities, err := s.entityRepo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	if len(entities) == 0 {
		return nil, fmt.Errorf("no entities found for project")
	}

	var allFiles []GeneratedFile

	// Generate code for each entity
	for _, entity := range entities {
		response, err := s.GenerateEntity(ctx, entity.ID, outputDir)
		if err != nil {
			return nil, fmt.Errorf("failed to generate entity %s: %w", entity.Name, err)
		}
		allFiles = append(allFiles, response.Files...)
	}

	return &GenerateCodeResponse{
		Files:   allFiles,
		Success: true,
		Message: fmt.Sprintf("Successfully generated %d files for project %s", len(allFiles), project.Name),
	}, nil
}

// PreviewEntity generates code preview without writing files
func (s *GeneratorService) PreviewEntity(ctx context.Context, entityID int64) (*GenerateCodeResponse, error) {
	return s.GenerateEntity(ctx, entityID, "")
}

// GetGeneratedFilesList returns list of files that will be generated
func (s *GeneratorService) GetGeneratedFilesList(ctx context.Context, entityID int64) ([]string, error) {
	entity, err := s.entityRepo.GetByID(entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	return s.generator.GetGeneratedFiles(entity.Name), nil
}

// GenerateEntityByUUID generates code for a specific entity by UUID
func (s *GeneratorService) GenerateEntityByUUID(ctx context.Context, entityUUID string, outputDir string) (*GenerateCodeResponse, error) {
	// Get entity by UUID
	entity, err := s.entityRepo.GetByUUID(entityUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	return s.GenerateEntity(ctx, entity.ID, outputDir)
}

// GenerateProjectByUUID generates code for all entities in a project by UUID
func (s *GeneratorService) GenerateProjectByUUID(ctx context.Context, projectUUID string, outputDir string) (*GenerateCodeResponse, error) {
	// Get project by UUID
	project, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return s.GenerateProject(ctx, project.ID, outputDir)
}

// PreviewEntityByUUID generates code preview without writing files by UUID
func (s *GeneratorService) PreviewEntityByUUID(ctx context.Context, entityUUID string) (*GenerateCodeResponse, error) {
	entity, err := s.entityRepo.GetByUUID(entityUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	return s.GenerateEntity(ctx, entity.ID, "")
}

// GetGeneratedFilesListByUUID returns list of files that will be generated by UUID
func (s *GeneratorService) GetGeneratedFilesListByUUID(ctx context.Context, entityUUID string) ([]string, error) {
	entity, err := s.entityRepo.GetByUUID(entityUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	return s.generator.GetGeneratedFiles(entity.Name), nil
}

// fetchRelationsForEntity fetches relations for an entity and converts to RelationContext
func (s *GeneratorService) fetchRelationsForEntity(entity *models.Entity) ([]generator.RelationContext, error) {
var relationContexts []generator.RelationContext

// Get relations where entity is source
sourceRelations, err := s.relationRepo.GetBySourceEntity(entity.ID)
if err != nil {
return nil, fmt.Errorf("failed to get source relations: %w", err)
}

// Get relations where entity is target
targetRelations, err := s.relationRepo.GetByTargetEntity(entity.ID)
if err != nil {
return nil, fmt.Errorf("failed to get target relations: %w", err)
}

// Convert source relations
for _, rel := range sourceRelations {
// For belongsTo: current entity has FK
if rel.RelationType == "belongsTo" {
relationContexts = append(relationContexts, generator.RelationContext{
SourceEntityName: rel.SourceEntityName,
SourceTableName:  generator.ToSnakeCase(rel.SourceEntityName),
TargetEntityName: rel.TargetEntityName,
TargetTableName:  generator.ToSnakeCase(rel.TargetEntityName),
FieldName:        rel.SourceFieldName,
RelationType:     rel.RelationType,
OnDelete:         rel.OnDelete,
OnUpdate:         rel.OnUpdate,
IsSource:         true,
})
}
}

// Convert target relations
for _, rel := range targetRelations {
// For hasOne/hasMany: current entity (target) has FK to source
if rel.RelationType == "hasOne" || rel.RelationType == "hasMany" {
relationContexts = append(relationContexts, generator.RelationContext{
SourceEntityName: rel.SourceEntityName,
SourceTableName:  generator.ToSnakeCase(rel.SourceEntityName),
TargetEntityName: rel.TargetEntityName,
TargetTableName:  generator.ToSnakeCase(rel.TargetEntityName),
FieldName:        rel.SourceFieldName,
RelationType:     rel.RelationType,
OnDelete:         rel.OnDelete,
OnUpdate:         rel.OnUpdate,
IsSource:         false,
})
}
}

return relationContexts, nil
}
