package models

import (
	"database/sql"
	"encoding/json"
	"errors"
)

// Relation represents a database relation between entities
type Relation struct {
	BaseEntity
	SourceEntityID  int64          `db:"source_entity_id" json:"-"`
	SourceFieldName string         `db:"source_field_name" json:"source_field_name"`
	TargetEntityID  int64          `db:"target_entity_id" json:"-"`
	RelationType    string         `db:"relation_type" json:"relation_type"`
	OnDelete        string         `db:"on_delete" json:"on_delete"`
	OnUpdate        string         `db:"on_update" json:"on_update"`
	JunctionTable   sql.NullString `db:"junction_table" json:"junction_table,omitempty"`
	Description     sql.NullString `db:"description" json:"description,omitempty"`
	Required        bool           `db:"required" json:"required"`

	// For JSON responses (populated from joins)
	SourceEntityUUID string `db:"-" json:"source_entity_id,omitempty"`
	TargetEntityUUID string `db:"-" json:"target_entity_id,omitempty"`
	SourceEntityName string `db:"-" json:"source_entity_name,omitempty"`
	TargetEntityName string `db:"-" json:"target_entity_name,omitempty"`
}

// Relation type constants
const (
	RelationTypeBelongsTo  = "belongsTo"
	RelationTypeHasOne     = "hasOne"
	RelationTypeHasMany    = "hasMany"
	RelationTypeManyToMany = "manyToMany"
)

// TableName returns the table name for this model
func (Relation) TableName() string {
	return "relations"
}

// Validate performs validation on the relation
func (r *Relation) Validate() error {
	if r.SourceEntityID == 0 {
		return errors.New("source_entity_id is required")
	}
	if r.SourceFieldName == "" {
		return errors.New("source_field_name is required")
	}
	if r.TargetEntityID == 0 {
		return errors.New("target_entity_id is required")
	}
	if r.RelationType == "" {
		return errors.New("relation_type is required")
	}

	// Validate relation type
	validTypes := map[string]bool{
		"belongsTo":   true,
		"hasOne":      true,
		"hasMany":     true,
		"manyToMany":  true,
	}
	if !validTypes[r.RelationType] {
		return errors.New("invalid relation_type: must be belongsTo, hasOne, hasMany, or manyToMany")
	}

	// Validate ON DELETE
	validOnDelete := map[string]bool{
		"CASCADE":   true,
		"SET NULL":  true,
		"RESTRICT":  true,
		"NO ACTION": true,
	}
	if r.OnDelete != "" && !validOnDelete[r.OnDelete] {
		return errors.New("invalid on_delete: must be CASCADE, SET NULL, RESTRICT, or NO ACTION")
	}

	// Validate ON UPDATE
	validOnUpdate := map[string]bool{
		"CASCADE":   true,
		"SET NULL":  true,
		"RESTRICT":  true,
		"NO ACTION": true,
	}
	if r.OnUpdate != "" && !validOnUpdate[r.OnUpdate] {
		return errors.New("invalid on_update: must be CASCADE, SET NULL, RESTRICT, or NO ACTION")
	}

	// Prevent self-reference for now (can be enabled later if needed)
	if r.SourceEntityID == r.TargetEntityID {
		return errors.New("self-referencing relations not supported yet")
	}

	return nil
}

// MarshalJSON customizes JSON output to expose UUID as "id"
func (r Relation) MarshalJSON() ([]byte, error) {
	type Alias Relation
	return json.Marshal(&struct {
		ID string `json:"id"`
		*Alias
	}{
		ID:    r.UUID,
		Alias: (*Alias)(&r),
	})
}

// RelationWithEntities includes entity names for display
type RelationWithEntities struct {
	Relation
	SourceEntity Entity `json:"source_entity"`
	TargetEntity Entity `json:"target_entity"`
}
