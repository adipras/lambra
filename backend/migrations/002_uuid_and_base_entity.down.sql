-- Rollback to previous schema with BIGINT IDs
-- This will drop UUID tables and recreate with BIGINT

-- Drop UUID tables
DROP TABLE IF EXISTS deployment_logs;
DROP TABLE IF EXISTS deployments;
DROP TABLE IF EXISTS generation_snapshots;
DROP TABLE IF EXISTS endpoints;
DROP TABLE IF EXISTS entities;
ALTER TABLE projects DROP FOREIGN KEY IF EXISTS fk_projects_git_repo;
DROP TABLE IF EXISTS git_repositories;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS projects;

-- Recreate with BIGINT IDs (original schema from 001)
-- Run the original 001_initial_schema.up.sql content
-- (This is just a placeholder - in production you'd include the full original schema)

-- For now, just indicate that a full rollback is needed
-- Users should reapply migration 001 after rolling back this one
