package generator

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yourusername/lambra/internal/models"
)

func TestCodeGenerator_PrepareContext(t *testing.T) {
	gen := NewCodeGenerator()

	// Create test project
	project := &models.Project{
		BaseEntity: models.BaseEntity{
			ID: 1,
		},
		Name:        "TestProject",
		Description: sql.NullString{String: "Test Description", Valid: true},
	}

	// Create test entity with fields
	fields := []models.EntityField{
		{
			Name:        "name",
			Type:        "string",
			Required:    true,
			Description: "User name",
			Length:      100,
		},
		{
			Name:        "email",
			Type:        "string",
			Required:    true,
			Description: "User email",
			Length:      255,
		},
		{
			Name:        "age",
			Type:        "int",
			Required:    false,
			Description: "User age",
		},
		{
			Name:        "created_date",
			Type:        "datetime",
			Required:    true,
			Description: "Creation date",
		},
	}

	fieldsJSON, _ := json.Marshal(fields)

	entity := &models.Entity{
		BaseEntity: models.BaseEntity{
			ID: 1,
		},
		ProjectID: 1,
		Name:      "User",
		TableName: "users",
		Fields:    fieldsJSON,
	}

	// Test PrepareContext
	ctx, err := gen.PrepareContext(project, entity)
	if err != nil {
		t.Fatalf("PrepareContext() error = %v", err)
	}

	// Verify context
	if ctx.EntityName != "User" {
		t.Errorf("EntityName = %v, want User", ctx.EntityName)
	}

	if ctx.EntityNameLC != "user" {
		t.Errorf("EntityNameLC = %v, want user", ctx.EntityNameLC)
	}

	if ctx.TableName != "users" {
		t.Errorf("TableName = %v, want users", ctx.TableName)
	}

	if len(ctx.Fields) != 4 {
		t.Errorf("Fields length = %v, want 4", len(ctx.Fields))
	}

	// Verify field conversion
	nameField := ctx.Fields[0]
	if nameField.Name != "Name" {
		t.Errorf("Field Name = %v, want Name", nameField.Name)
	}
	if nameField.GoType != "string" {
		t.Errorf("Field GoType = %v, want string", nameField.GoType)
	}
	if nameField.Required != true {
		t.Errorf("Field Required = %v, want true", nameField.Required)
	}

	// Verify optional field
	ageField := ctx.Fields[2]
	if ageField.GoType != "*int64" {
		t.Errorf("Age GoType = %v, want *int64", ageField.GoType)
	}

	// Verify datetime field
	dateField := ctx.Fields[3]
	if dateField.GoType != "time.Time" {
		t.Errorf("Date GoType = %v, want time.Time", dateField.GoType)
	}
}

func TestCodeGenerator_ValidateContext(t *testing.T) {
	gen := NewCodeGenerator()

	tests := []struct {
		name    string
		ctx     *GenerateContext
		wantErr bool
	}{
		{
			name: "valid context",
			ctx: &GenerateContext{
				EntityName: "User",
				TableName:  "users",
				Fields: []FieldContext{
					{Name: "Name", Type: "string"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing entity name",
			ctx: &GenerateContext{
				TableName: "users",
				Fields: []FieldContext{
					{Name: "Name", Type: "string"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing table name",
			ctx: &GenerateContext{
				EntityName: "User",
				Fields: []FieldContext{
					{Name: "Name", Type: "string"},
				},
			},
			wantErr: true,
		},
		{
			name: "no fields",
			ctx: &GenerateContext{
				EntityName: "User",
				TableName:  "users",
				Fields:     []FieldContext{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.ValidateContext(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCodeGenerator_GenerateModel(t *testing.T) {
	gen := NewCodeGenerator()

	ctx := &GenerateContext{
		EntityName:   "User",
		EntityNameLC: "user",
		TableName:    "users",
		PackageName:  "models",
		Imports:      []string{"time"},
		Fields: []FieldContext{
			{
				Name:        "Name",
				NameLC:      "name",
				Type:        "string",
				GoType:      "string",
				JSONTag:     `json:"name"`,
				DBTag:       `db:"name"`,
				ValidateTag: `validate:"required"`,
				Required:    true,
				Description: "User name",
			},
			{
				Name:        "Email",
				NameLC:      "email",
				Type:        "string",
				GoType:      "string",
				JSONTag:     `json:"email"`,
				DBTag:       `db:"email"`,
				ValidateTag: `validate:"required,email"`,
				Required:    true,
				Description: "User email",
			},
		},
	}

	code, err := gen.GenerateModel(ctx)
	if err != nil {
		t.Fatalf("GenerateModel() error = %v", err)
	}

	// Verify generated code contains expected elements
	expectedStrings := []string{
		"package models",
		"type User struct",
		"BaseEntity",
		"Name string",
		"Email string",
		`json:"name"`,
		`json:"email"`,
		`db:"name"`,
		`db:"email"`,
		"func (user *User) TableName() string",
		`return "users"`,
		"func (user *User) Validate() error",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(code, expected) {
			t.Errorf("Generated model does not contain: %q", expected)
		}
	}
}

func TestCodeGenerator_GenerateRepository(t *testing.T) {
	gen := NewCodeGenerator()

	ctx := &GenerateContext{
		EntityName:   "User",
		EntityNameLC: "user",
		TableName:    "users",
		PackageName:  "repository",
		Fields: []FieldContext{
			{
				Name:    "Name",
				NameLC:  "name",
				Type:    "string",
				GoType:  "string",
				JSONTag: `json:"name"`,
				DBTag:   `db:"name"`,
			},
		},
	}

	code, err := gen.GenerateRepository(ctx)
	if err != nil {
		t.Fatalf("GenerateRepository() error = %v", err)
	}

	// Verify generated repository contains expected methods
	expectedStrings := []string{
		"package repository",
		"type UserRepository struct",
		"func NewUserRepository",
		"func (r *UserRepository) Create",
		"func (r *UserRepository) GetByID",
		"func (r *UserRepository) GetByUUID",
		"func (r *UserRepository) List",
		"func (r *UserRepository) Update",
		"func (r *UserRepository) Delete",
		"func (r *UserRepository) Count",
		"INSERT INTO users",
		"SELECT * FROM users",
		"UPDATE users",
		"deleted_at IS NULL",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(code, expected) {
			t.Errorf("Generated repository does not contain: %q", expected)
		}
	}
}

func TestCodeGenerator_GenerateService(t *testing.T) {
	gen := NewCodeGenerator()

	ctx := &GenerateContext{
		EntityName:   "User",
		EntityNameLC: "user",
		TableName:    "users",
		Fields:       []FieldContext{},
	}

	code, err := gen.GenerateService(ctx)
	if err != nil {
		t.Fatalf("GenerateService() error = %v", err)
	}

	expectedStrings := []string{
		"package service",
		"type UserService struct",
		"func NewUserService",
		"func (s *UserService) Create",
		"func (s *UserService) GetByID",
		"func (s *UserService) GetByUUID",
		"func (s *UserService) List",
		"func (s *UserService) Update",
		"func (s *UserService) Delete",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(code, expected) {
			t.Errorf("Generated service does not contain: %q", expected)
		}
	}
}

func TestCodeGenerator_GenerateHandler(t *testing.T) {
	gen := NewCodeGenerator()

	ctx := &GenerateContext{
		EntityName:   "User",
		EntityNameLC: "user",
		Fields:       []FieldContext{},
	}

	code, err := gen.GenerateHandler(ctx)
	if err != nil {
		t.Fatalf("GenerateHandler() error = %v", err)
	}

	expectedStrings := []string{
		"package handlers",
		"type UserHandler struct",
		"func NewUserHandler",
		"func (h *UserHandler) CreateUser",
		"func (h *UserHandler) GetUser",
		"func (h *UserHandler) ListUsers",
		"func (h *UserHandler) UpdateUser",
		"func (h *UserHandler) DeleteUser",
		"c.JSON(",
		"http.Status",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(code, expected) {
			t.Errorf("Generated handler does not contain: %q", expected)
		}
	}
}

func TestCodeGenerator_GenerateDTO(t *testing.T) {
	gen := NewCodeGenerator()

	ctx := &GenerateContext{
		EntityName:   "User",
		EntityNameLC: "user",
		Fields: []FieldContext{
			{
				Name:    "Name",
				NameLC:  "name",
				GoType:  "string",
				JSONTag: `json:"name"`,
			},
		},
	}

	code, err := gen.GenerateDTO(ctx)
	if err != nil {
		t.Fatalf("GenerateDTO() error = %v", err)
	}

	expectedStrings := []string{
		"package dto",
		"type CreateUserRequest struct",
		"type UpdateUserRequest struct",
		"type UserResponse struct",
		"func (r CreateUserRequest) ToModel()",
		"func (r UpdateUserRequest) ToModel()",
		"func (r UserResponse) FromModel(",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(code, expected) {
			t.Errorf("Generated DTO does not contain: %q", expected)
		}
	}
}

func TestCodeGenerator_GenerateMigration(t *testing.T) {
	gen := NewCodeGenerator()

	ctx := &GenerateContext{
		EntityName: "User",
		TableName:  "users",
		Fields: []FieldContext{
			{
				Name:     "Name",
				NameLC:   "name",
				Type:     "string",
				GoType:   "string",
				Required: true,
			},
			{
				Name:     "Email",
				NameLC:   "email",
				Type:     "string",
				GoType:   "string",
				Required: true,
			},
		},
	}

	up, down, err := gen.GenerateMigration(ctx)
	if err != nil {
		t.Fatalf("GenerateMigration() error = %v", err)
	}

	// Check UP migration
	upExpected := []string{
		"CREATE TABLE IF NOT EXISTS users",
		"id BIGSERIAL PRIMARY KEY",
		"uuid UUID NOT NULL UNIQUE",
		"created_at TIMESTAMP",
		"updated_at TIMESTAMP",
		"deleted_at TIMESTAMP",
		"CREATE INDEX idx_users_uuid",
		"CREATE INDEX idx_users_deleted_at",
	}

	for _, expected := range upExpected {
		if !strings.Contains(up, expected) {
			t.Errorf("UP migration does not contain: %q", expected)
		}
	}

	// Check DOWN migration
	if !strings.Contains(down, "DROP TABLE IF EXISTS users") {
		t.Errorf("DOWN migration does not contain DROP TABLE")
	}
}

func TestCodeGenerator_GetGeneratedFiles(t *testing.T) {
	gen := NewCodeGenerator()

	files := gen.GetGeneratedFiles("User")

	expectedFiles := []string{
		"models/user.go",
		"repository/user_repository.go",
		"service/user_service.go",
		"api/handlers/user_handler.go",
		"api/dto/user_dto.go",
	}

	if len(files) != len(expectedFiles) {
		t.Errorf("GetGeneratedFiles() returned %d files, want %d", len(files), len(expectedFiles))
	}

	for _, expected := range expectedFiles {
		found := false
		for _, file := range files {
			if file == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetGeneratedFiles() missing file: %s", expected)
		}
	}
}
