package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yourusername/lambra/internal/models"
)

func TestGenerateMigrationSQL_WithUniqueFields(t *testing.T) {
	// Create deployment service (we only need the method, not full dependencies)
	service := &DeploymentService{}

	// Create entity with fields that have unique constraints
	fields := []models.EntityField{
		{
			Name:     "sku",
			Type:     "string",
			Length:   10,
			Unique:   true,
			Required: true,
		},
		{
			Name:     "name",
			Type:     "string",
			Unique:   true,
			Required: true,
		},
		{
			Name:     "is_active",
			Type:     "bool",
			Unique:   false,
			Required: true,
		},
	}

	fieldsJSON, _ := json.Marshal(fields)

	entity := models.Entity{
		Name:      "Product",
		TableName: "products",
		Fields:    fieldsJSON,
	}

	// Generate migration SQL
	sql := service.generateMigrationSQL(entity, []models.Relation{})

	// Verify UNIQUE constraint is added for unique fields
	if !strings.Contains(sql, "sku VARCHAR(10) NOT NULL UNIQUE") {
		t.Errorf("Migration SQL should contain UNIQUE constraint for sku field")
	}

	if !strings.Contains(sql, "name VARCHAR(255) NOT NULL UNIQUE") {
		t.Errorf("Migration SQL should contain UNIQUE constraint for name field")
	}

	// Verify UNIQUE constraint is NOT added for non-unique fields
	if strings.Contains(sql, "is_active BOOLEAN NOT NULL UNIQUE") {
		t.Errorf("Migration SQL should NOT contain UNIQUE constraint for is_active field")
	}

	// But should have the field without UNIQUE
	if !strings.Contains(sql, "is_active BOOLEAN NOT NULL") {
		t.Errorf("Migration SQL should contain is_active field")
	}

	t.Logf("Generated SQL:\n%s", sql)
}

func TestGenerateMigrationSQL_WithoutUniqueFields(t *testing.T) {
	service := &DeploymentService{}

	fields := []models.EntityField{
		{
			Name:     "title",
			Type:     "string",
			Unique:   false,
			Required: true,
		},
		{
			Name:     "count",
			Type:     "int",
			Unique:   false,
			Required: false,
		},
	}

	fieldsJSON, _ := json.Marshal(fields)

	entity := models.Entity{
		Name:      "Item",
		TableName: "items",
		Fields:    fieldsJSON,
	}

	sql := service.generateMigrationSQL(entity, []models.Relation{})

	// Verify no UNIQUE constraints are added
	if strings.Contains(sql, "title VARCHAR(255) NOT NULL UNIQUE") {
		t.Errorf("Migration SQL should NOT contain UNIQUE constraint for title field")
	}

	if strings.Contains(sql, "count INT UNIQUE") {
		t.Errorf("Migration SQL should NOT contain UNIQUE constraint for count field")
	}

	// Fields should still be present
	if !strings.Contains(sql, "title VARCHAR(255) NOT NULL") {
		t.Errorf("Migration SQL should contain title field")
	}
}
