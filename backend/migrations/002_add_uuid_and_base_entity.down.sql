-- Rollback: Remove UUID and base entity fields

SET FOREIGN_KEY_CHECKS = 0;

-- Projects
ALTER TABLE projects
  DROP COLUMN uuid,
  DROP COLUMN created_by,
  DROP COLUMN updated_by,
  DROP COLUMN deleted_by,
  DROP COLUMN deleted_at,
  DROP INDEX idx_uuid,
  DROP INDEX idx_deleted_at,
  MODIFY COLUMN id BIGINT AUTO_INCREMENT;

-- Git Repositories
ALTER TABLE git_repositories
  DROP COLUMN uuid,
  DROP COLUMN created_by,
  DROP COLUMN updated_by,
  DROP COLUMN deleted_by,
  DROP COLUMN deleted_at,
  DROP INDEX idx_uuid,
  DROP INDEX idx_deleted_at,
  MODIFY COLUMN id BIGINT AUTO_INCREMENT;

-- Entities
ALTER TABLE entities
  DROP COLUMN uuid,
  DROP COLUMN created_by,
  DROP COLUMN updated_by,
  DROP COLUMN deleted_by,
  DROP COLUMN deleted_at,
  DROP INDEX idx_uuid,
  DROP INDEX idx_deleted_at,
  MODIFY COLUMN id BIGINT AUTO_INCREMENT;

-- Endpoints
ALTER TABLE endpoints
  DROP COLUMN uuid,
  DROP COLUMN created_by,
  DROP COLUMN updated_by,
  DROP COLUMN deleted_by,
  DROP COLUMN deleted_at,
  DROP INDEX idx_uuid,
  DROP INDEX idx_deleted_at,
  MODIFY COLUMN id BIGINT AUTO_INCREMENT;

-- Generation Snapshots
ALTER TABLE generation_snapshots
  DROP COLUMN uuid,
  DROP COLUMN updated_by,
  DROP COLUMN deleted_by,
  DROP COLUMN deleted_at,
  DROP COLUMN updated_at,
  DROP INDEX idx_uuid,
  DROP INDEX idx_deleted_at,
  MODIFY COLUMN id BIGINT AUTO_INCREMENT;

-- Deployments
ALTER TABLE deployments
  DROP COLUMN uuid,
  DROP COLUMN created_by,
  DROP COLUMN updated_by,
  DROP COLUMN deleted_by,
  DROP COLUMN deleted_at,
  DROP INDEX idx_uuid,
  DROP INDEX idx_deleted_at,
  MODIFY COLUMN id BIGINT AUTO_INCREMENT;

-- Deployment Logs
ALTER TABLE deployment_logs
  MODIFY COLUMN id BIGINT AUTO_INCREMENT;

-- Templates
ALTER TABLE templates
  DROP COLUMN uuid,
  DROP COLUMN created_by,
  DROP COLUMN updated_by,
  DROP COLUMN deleted_by,
  DROP COLUMN deleted_at,
  DROP INDEX idx_uuid,
  DROP INDEX idx_deleted_at,
  MODIFY COLUMN id BIGINT AUTO_INCREMENT;

SET FOREIGN_KEY_CHECKS = 1;
