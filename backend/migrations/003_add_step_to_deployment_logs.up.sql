-- Add step column to deployment_logs for tracking deployment progress
ALTER TABLE deployment_logs ADD COLUMN step VARCHAR(50) DEFAULT NULL AFTER message;
