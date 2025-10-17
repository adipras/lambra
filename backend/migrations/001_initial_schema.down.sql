-- Rollback: Drop all tables in reverse order (respecting foreign key dependencies)

SET FOREIGN_KEY_CHECKS = 0;

-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS deployment_logs;
DROP TABLE IF EXISTS deployments;
DROP TABLE IF EXISTS generation_snapshots;
DROP TABLE IF EXISTS endpoints;
DROP TABLE IF EXISTS entities;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS git_repositories;
DROP TABLE IF EXISTS projects;

SET FOREIGN_KEY_CHECKS = 1;
