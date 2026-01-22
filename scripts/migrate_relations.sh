#!/bin/bash

# Migration Script: Convert Legacy Field-based Relations to Relations Table
# This script is OPTIONAL - only needed if you have existing projects with relation fields
# 
# Usage: ./migrate_relations.sh
# 
# Prerequisites:
# - Docker services must be running (make up)
# - Database must be accessible
# - Backup your database before running!

set -e

echo "=========================================="
echo "Legacy Relations Migration Script"
echo "=========================================="
echo ""
echo "⚠️  WARNING: This will convert field-based relations to relations table"
echo "⚠️  Make sure you have backed up your database!"
echo ""
read -p "Do you want to continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Migration cancelled."
    exit 0
fi

echo ""
echo "Starting migration..."
echo ""

# Run migration SQL
docker-compose exec -T mysql mysql -ulambra -plambra_secret lambra_db << 'EOF'

-- Migration: Convert legacy field-based relations to relations table
-- This is a one-time migration for existing projects

START TRANSACTION;

-- Create temporary table to hold legacy relations
CREATE TEMPORARY TABLE temp_legacy_relations AS
SELECT 
    e.id as source_entity_id,
    e.project_id,
    JSON_UNQUOTE(JSON_EXTRACT(field_data, '$.name')) as field_name,
    JSON_UNQUOTE(JSON_EXTRACT(field_data, '$.relation_type')) as relation_type,
    JSON_UNQUOTE(JSON_EXTRACT(field_data, '$.related_entity')) as related_entity_name,
    JSON_UNQUOTE(JSON_EXTRACT(field_data, '$.on_delete')) as on_delete,
    JSON_UNQUOTE(JSON_EXTRACT(field_data, '$.foreign_key')) as foreign_key,
    JSON_EXTRACT(field_data, '$.required') as is_required
FROM entities e,
JSON_TABLE(
    e.fields,
    '$[*]' COLUMNS(
        field_data JSON PATH '$'
    )
) as jt
WHERE JSON_UNQUOTE(JSON_EXTRACT(field_data, '$.type')) = 'relation'
  AND e.deleted_at IS NULL;

-- Display what will be migrated
SELECT 
    COUNT(*) as legacy_relations_found,
    GROUP_CONCAT(DISTINCT relation_type) as relation_types
FROM temp_legacy_relations;

-- Insert into relations table
-- Note: This creates new relations, does NOT modify existing entity fields
INSERT INTO relations (
    id,
    uuid,
    source_entity_id,
    source_field_name,
    target_entity_id,
    relation_type,
    on_delete,
    on_update,
    junction_table,
    description,
    required,
    created_by,
    created_at,
    updated_at
)
SELECT
    -- Generate ID from timestamp + random
    FLOOR(UNIX_TIMESTAMP(NOW()) * 1000000 + FLOOR(RAND() * 1000000)),
    -- Generate UUID
    UUID(),
    -- Source entity
    tlr.source_entity_id,
    -- Field name (FK column name)
    COALESCE(
        tlr.foreign_key,
        CONCAT(LOWER(tlr.related_entity_name), '_id')
    ),
    -- Target entity (lookup by name)
    (SELECT id FROM entities WHERE name = tlr.related_entity_name AND project_id = tlr.project_id LIMIT 1),
    -- Relation type
    tlr.relation_type,
    -- ON DELETE behavior
    COALESCE(tlr.on_delete, 'RESTRICT'),
    -- ON UPDATE behavior (default)
    'CASCADE',
    -- Junction table (for manyToMany)
    CASE 
        WHEN tlr.relation_type = 'manyToMany' THEN
            CONCAT(
                (SELECT table_name FROM entities WHERE id = tlr.source_entity_id),
                '_',
                (SELECT table_name FROM entities WHERE name = tlr.related_entity_name AND project_id = tlr.project_id LIMIT 1)
            )
        ELSE NULL
    END,
    -- Description
    CONCAT('Migrated from legacy field: ', tlr.field_name),
    -- Required flag
    CASE WHEN tlr.is_required = 'true' THEN 1 ELSE 0 END,
    -- Audit fields
    'migration_script',
    NOW(),
    NOW()
FROM temp_legacy_relations tlr
WHERE 
    -- Only if target entity exists
    (SELECT id FROM entities WHERE name = tlr.related_entity_name AND project_id = tlr.project_id LIMIT 1) IS NOT NULL
    -- And relation doesn't already exist in new format
    AND NOT EXISTS (
        SELECT 1 FROM relations r
        WHERE r.source_entity_id = tlr.source_entity_id
          AND r.target_entity_id = (SELECT id FROM entities WHERE name = tlr.related_entity_name AND project_id = tlr.project_id LIMIT 1)
          AND r.deleted_at IS NULL
    );

-- Show results
SELECT 
    COUNT(*) as relations_migrated
FROM relations
WHERE created_by = 'migration_script';

-- Note: We do NOT remove relation fields from entities.fields JSON
-- They will be displayed as "legacy" in the diagram with dashed lines
-- The new relations will take precedence

COMMIT;

SELECT '✅ Migration completed successfully!' as status;

EOF

echo ""
echo "=========================================="
echo "Migration completed!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Check the migration results above"
echo "2. Open Lambra UI and go to Diagram View"
echo "3. Verify that relations are displayed correctly"
echo "4. Legacy relations will show as dashed lines"
echo "5. New relations from relations table will show as solid lines"
echo ""
echo "Note: Entity fields JSON has NOT been modified."
echo "Legacy relation fields are preserved for backward compatibility."
echo ""
