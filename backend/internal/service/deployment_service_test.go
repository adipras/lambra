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

func TestGenerateMigrationSQL_WithHasManyRelations(t *testing.T) {
service := &DeploymentService{}

// Create Stock entity
stockFields := []models.EntityField{
{
Name:     "quantity",
Type:     "int",
Required: true,
},
}
stockFieldsJSON, _ := json.Marshal(stockFields)

stockEntity := models.Entity{
BaseEntity: models.BaseEntity{
ID: 3,
},
Name:      "Stock",
TableName: "stocks",
Fields:    stockFieldsJSON,
}

// Create relations: Product hasMany Stock, Location hasMany Stock
relations := []models.Relation{
{
BaseEntity: models.BaseEntity{
ID: 1,
},
SourceEntityID:   1, // Product
TargetEntityID:   3, // Stock
SourceEntityName: "Product",
TargetEntityName: "Stock",
RelationType:     models.RelationTypeHasMany,
SourceFieldName:  "product_id",
			Required:         true,
OnDelete:         "CASCADE",
OnUpdate:         "CASCADE",
},
{
BaseEntity: models.BaseEntity{
ID: 2,
},
SourceEntityID:   2, // Location
TargetEntityID:   3, // Stock
SourceEntityName: "Location",
TargetEntityName: "Stock",
RelationType:     models.RelationTypeHasMany,
SourceFieldName:  "location_id",
			Required:         true,
OnDelete:         "CASCADE",
OnUpdate:         "CASCADE",
},
}

sql := service.generateMigrationSQL(stockEntity, relations)

// Verify FK columns are added
if !strings.Contains(sql, "product_id BIGINT NOT NULL") {
t.Errorf("SQL should contain product_id FK column, got:\n%s", sql)
}

if !strings.Contains(sql, "location_id BIGINT NOT NULL") {
t.Errorf("SQL should contain location_id FK column, got:\n%s", sql)
}

// Verify FK constraints are added
if !strings.Contains(sql, "FOREIGN KEY (product_id)") {
t.Errorf("SQL should contain FK constraint for product_id, got:\n%s", sql)
}

if !strings.Contains(sql, "FOREIGN KEY (location_id)") {
t.Errorf("SQL should contain FK constraint for location_id, got:\n%s", sql)
}

// Verify ON DELETE and ON UPDATE
if !strings.Contains(sql, "ON DELETE CASCADE") {
t.Errorf("SQL should contain ON DELETE CASCADE, got:\n%s", sql)
}

if !strings.Contains(sql, "ON UPDATE CASCADE") {
t.Errorf("SQL should contain ON UPDATE CASCADE, got:\n%s", sql)
}

t.Logf("Generated SQL:\n%s", sql)
}

func TestGenerateMigrationSQL_WithBelongsToRelation(t *testing.T) {
service := &DeploymentService{}

// Create Post entity with belongsTo User
postFields := []models.EntityField{
{
Name:     "title",
Type:     "string",
Required: true,
},
}
postFieldsJSON, _ := json.Marshal(postFields)

postEntity := models.Entity{
BaseEntity: models.BaseEntity{
ID: 2,
},
Name:      "Post",
TableName: "posts",
Fields:    postFieldsJSON,
}

// Post belongsTo User
relations := []models.Relation{
{
BaseEntity: models.BaseEntity{
ID: 1,
},
SourceEntityID:   2, // Post
TargetEntityID:   1, // User
SourceEntityName: "Post",
TargetEntityName: "User",
RelationType:     models.RelationTypeBelongsTo,
SourceFieldName:  "user_id",
			Required:         true,
OnDelete:         "CASCADE",
OnUpdate:         "CASCADE",
},
}

sql := service.generateMigrationSQL(postEntity, relations)

// Verify FK column is added
if !strings.Contains(sql, "user_id BIGINT NOT NULL") {
t.Errorf("SQL should contain user_id FK column, got:\n%s", sql)
}

// Verify FK constraint is added
if !strings.Contains(sql, "FOREIGN KEY (user_id)") {
t.Errorf("SQL should contain FK constraint for user_id, got:\n%s", sql)
}

t.Logf("Generated SQL:\n%s", sql)
}

func TestGenerateMigrationSQL_NoRelations(t *testing.T) {
service := &DeploymentService{}

// Create Product entity without relations
productFields := []models.EntityField{
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
}
productFieldsJSON, _ := json.Marshal(productFields)

productEntity := models.Entity{
BaseEntity: models.BaseEntity{
ID: 1,
},
Name:      "Product",
TableName: "products",
Fields:    productFieldsJSON,
}

sql := service.generateMigrationSQL(productEntity, []models.Relation{})

// Verify basic table structure
if !strings.Contains(sql, "CREATE TABLE IF NOT EXISTS products") {
t.Errorf("SQL should contain CREATE TABLE")
}

if !strings.Contains(sql, "sku VARCHAR(10) NOT NULL UNIQUE") {
t.Errorf("SQL should contain sku with UNIQUE constraint")
}

if !strings.Contains(sql, "name VARCHAR(255) NOT NULL UNIQUE") {
t.Errorf("SQL should contain name with UNIQUE constraint")
}

// Should NOT contain any FK
if strings.Contains(sql, "FOREIGN KEY") {
t.Errorf("SQL should NOT contain FOREIGN KEY when no relations, got:\n%s", sql)
}

t.Logf("Generated SQL:\n%s", sql)
}

func TestGenerateMigrationSQL_NoDuplicateFields(t *testing.T) {
service := &DeploymentService{}

// Stock entity has product_id as regular field (user mistake)
// AND relation also adds product_id (should skip the regular field)
stockFields := []models.EntityField{
{
Name:     "product_id",  // User accidentally created this field
Type:     "int",
Required: true,
},
{
Name:     "quantity",
Type:     "int",
Required: true,
},
}
stockFieldsJSON, _ := json.Marshal(stockFields)

stockEntity := models.Entity{
BaseEntity: models.BaseEntity{
ID: 3,
},
Name:      "Stock",
TableName: "stocks",
Fields:    stockFieldsJSON,
}

// Product hasMany Stock with product_id
relations := []models.Relation{
{
SourceEntityID:   1,
TargetEntityID:   3,
SourceEntityName: "Product",
TargetEntityName: "Stock",
RelationType:     models.RelationTypeHasMany,
SourceFieldName:  "product_id",
OnDelete:         "CASCADE",
OnUpdate:         "CASCADE",
Required:         true,
},
}

sql := service.generateMigrationSQL(stockEntity, relations)

// Count occurrences of product_id
productIDCount := strings.Count(sql, "product_id")

// Should appear exactly 2 times:
// 1. Column definition: product_id BIGINT NOT NULL
// 2. FK constraint: FOREIGN KEY (product_id)
if productIDCount != 2 {
t.Errorf("product_id should appear exactly 2 times (column + FK), found %d times in:\n%s", productIDCount, sql)
}

// Should NOT have INT type (from regular field)
if strings.Contains(sql, "product_id INT") {
t.Errorf("SQL should not contain 'product_id INT' (regular field should be skipped), got:\n%s", sql)
}

// Should have BIGINT type (from FK)
if !strings.Contains(sql, "product_id BIGINT NOT NULL") {
t.Errorf("SQL should contain 'product_id BIGINT NOT NULL' (from FK), got:\n%s", sql)
}

t.Logf("Generated SQL:\n%s", sql)
}
