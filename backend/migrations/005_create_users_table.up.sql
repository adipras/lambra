-- Create users table for platform authentication
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) NOT NULL UNIQUE,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'admin',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    created_by VARCHAR(50) NULL,
    updated_by VARCHAR(50) NULL,
    INDEX idx_users_uuid (uuid),
    INDEX idx_users_deleted_at (deleted_at)
);

-- Seed default admin user (username: admin, password: admin123)
INSERT INTO users (uuid, username, password_hash, name, role)
VALUES (
    UUID(),
    'admin',
    '$2a$10$vImqR23ugSwjucz4w9fJzuJTmiu8R6vN8vUVzij5gW3kE7XzSsLEy',
    'Administrator',
    'admin'
)
ON DUPLICATE KEY UPDATE username = username;
