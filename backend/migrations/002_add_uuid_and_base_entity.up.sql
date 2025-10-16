-- Migration to add UUID field and base entity audit fields
-- Strategy: ID (BIGINT generated) + UUID (CHAR36) for dual identifiers

SET FOREIGN_KEY_CHECKS = 0;

-- Projects table
ALTER TABLE projects
  MODIFY COLUMN id BIGINT NOT NULL,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (id),
  ADD COLUMN uuid CHAR(36) NOT NULL UNIQUE AFTER id,
  ADD COLUMN created_by VARCHAR(100) AFTER namespace,
  ADD COLUMN updated_by VARCHAR(100) AFTER created_by,
  ADD COLUMN deleted_by VARCHAR(100) AFTER updated_by,
  ADD COLUMN deleted_at TIMESTAMP NULL AFTER updated_at,
  ADD INDEX idx_uuid (uuid),
  ADD INDEX idx_deleted_at (deleted_at);

-- Git Repositories table
ALTER TABLE git_repositories
  MODIFY COLUMN id BIGINT NOT NULL,
  MODIFY COLUMN project_id BIGINT NOT NULL,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (id),
  ADD COLUMN uuid CHAR(36) NOT NULL UNIQUE AFTER id,
  ADD COLUMN created_by VARCHAR(100) AFTER last_commit_hash,
  ADD COLUMN updated_by VARCHAR(100) AFTER created_by,
  ADD COLUMN deleted_by VARCHAR(100) AFTER updated_by,
  ADD COLUMN deleted_at TIMESTAMP NULL AFTER updated_at,
  ADD INDEX idx_uuid (uuid),
  ADD INDEX idx_deleted_at (deleted_at);

-- Entities table
ALTER TABLE entities
  MODIFY COLUMN id BIGINT NOT NULL,
  MODIFY COLUMN project_id BIGINT NOT NULL,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (id),
  ADD COLUMN uuid CHAR(36) NOT NULL UNIQUE AFTER id,
  ADD COLUMN created_by VARCHAR(100) AFTER fields,
  ADD COLUMN updated_by VARCHAR(100) AFTER created_by,
  ADD COLUMN deleted_by VARCHAR(100) AFTER updated_by,
  ADD COLUMN deleted_at TIMESTAMP NULL AFTER updated_at,
  ADD INDEX idx_uuid (uuid),
  ADD INDEX idx_deleted_at (deleted_at);

-- Endpoints table
ALTER TABLE endpoints
  MODIFY COLUMN id BIGINT NOT NULL,
  MODIFY COLUMN entity_id BIGINT NOT NULL,
  MODIFY COLUMN project_id BIGINT NOT NULL,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (id),
  ADD COLUMN uuid CHAR(36) NOT NULL UNIQUE AFTER id,
  ADD COLUMN created_by VARCHAR(100) AFTER require_auth,
  ADD COLUMN updated_by VARCHAR(100) AFTER created_by,
  ADD COLUMN deleted_by VARCHAR(100) AFTER updated_by,
  ADD COLUMN deleted_at TIMESTAMP NULL AFTER updated_at,
  ADD INDEX idx_uuid (uuid),
  ADD INDEX idx_deleted_at (deleted_at);

-- Generation Snapshots table
ALTER TABLE generation_snapshots
  MODIFY COLUMN id BIGINT NOT NULL,
  MODIFY COLUMN project_id BIGINT NOT NULL,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (id),
  ADD COLUMN uuid CHAR(36) NOT NULL UNIQUE AFTER id,
  ADD COLUMN updated_by VARCHAR(100) AFTER created_by,
  ADD COLUMN deleted_by VARCHAR(100) AFTER updated_by,
  ADD COLUMN deleted_at TIMESTAMP NULL AFTER created_at,
  ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER deleted_at,
  ADD INDEX idx_uuid (uuid),
  ADD INDEX idx_deleted_at (deleted_at);

-- Deployments table
ALTER TABLE deployments
  MODIFY COLUMN id BIGINT NOT NULL,
  MODIFY COLUMN project_id BIGINT NOT NULL,
  MODIFY COLUMN snapshot_id BIGINT,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (id),
  ADD COLUMN uuid CHAR(36) NOT NULL UNIQUE AFTER id,
  ADD COLUMN created_by VARCHAR(100) AFTER created_at,
  ADD COLUMN updated_by VARCHAR(100) AFTER created_by,
  ADD COLUMN deleted_by VARCHAR(100) AFTER updated_by,
  ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER deleted_by,
  ADD COLUMN deleted_at TIMESTAMP NULL AFTER updated_at,
  ADD INDEX idx_uuid (uuid),
  ADD INDEX idx_deleted_at (deleted_at);

-- Deployment Logs table (no UUID needed, just audit fields)
ALTER TABLE deployment_logs
  MODIFY COLUMN id BIGINT NOT NULL,
  MODIFY COLUMN deployment_id BIGINT NOT NULL,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (id);

-- Templates table
ALTER TABLE templates
  MODIFY COLUMN id BIGINT NOT NULL,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (id),
  ADD COLUMN uuid CHAR(36) NOT NULL UNIQUE AFTER id,
  ADD COLUMN created_by VARCHAR(100) AFTER is_default,
  ADD COLUMN updated_by VARCHAR(100) AFTER created_by,
  ADD COLUMN deleted_by VARCHAR(100) AFTER updated_by,
  ADD COLUMN deleted_at TIMESTAMP NULL AFTER updated_at,
  ADD INDEX idx_uuid (uuid),
  ADD INDEX idx_deleted_at (deleted_at);

SET FOREIGN_KEY_CHECKS = 1;
