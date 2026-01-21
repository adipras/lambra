-- Create relations table for visual database diagram
CREATE TABLE IF NOT EXISTS relations (
    id BIGINT PRIMARY KEY,
    uuid CHAR(36) UNIQUE NOT NULL,
    
    -- Source entity (the one that has the foreign key)
    source_entity_id BIGINT NOT NULL,
    source_field_name VARCHAR(100) NOT NULL,
    
    -- Target entity (the one being referenced)
    target_entity_id BIGINT NOT NULL,
    
    -- Relation configuration
    relation_type VARCHAR(50) NOT NULL, -- belongsTo, hasOne, hasMany, manyToMany
    on_delete VARCHAR(50) DEFAULT 'RESTRICT', -- CASCADE, SET NULL, RESTRICT, NO ACTION
    on_update VARCHAR(50) DEFAULT 'CASCADE',  -- CASCADE, SET NULL, RESTRICT, NO ACTION
    
    -- Junction table name for manyToMany relations
    junction_table VARCHAR(100),
    
    -- Metadata
    description TEXT,
    
    -- Audit fields
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    deleted_by VARCHAR(100),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    
    -- Foreign keys
    CONSTRAINT fk_relation_source_entity FOREIGN KEY (source_entity_id) REFERENCES entities(id),
    CONSTRAINT fk_relation_target_entity FOREIGN KEY (target_entity_id) REFERENCES entities(id),
    
    -- Indexes
    INDEX idx_relation_uuid (uuid),
    INDEX idx_relation_source_entity (source_entity_id),
    INDEX idx_relation_target_entity (target_entity_id),
    INDEX idx_relation_deleted_at (deleted_at)
);
