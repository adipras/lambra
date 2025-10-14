-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS deployment_logs;
DROP TABLE IF EXISTS deployments;
DROP TABLE IF EXISTS generation_snapshots;
DROP TABLE IF EXISTS endpoints;
DROP TABLE IF EXISTS entities;

-- Remove foreign key from projects before dropping git_repositories
ALTER TABLE projects DROP FOREIGN KEY IF EXISTS fk_projects_git_repo;
ALTER TABLE projects DROP COLUMN IF EXISTS git_repo_id;

DROP TABLE IF EXISTS git_repositories;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS projects;
