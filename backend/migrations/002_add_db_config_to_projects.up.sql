-- Add database configuration columns to projects table
ALTER TABLE projects
    ADD COLUMN db_host VARCHAR(255) NOT NULL DEFAULT 'localhost' AFTER namespace,
    ADD COLUMN db_port INT NOT NULL DEFAULT 3306 AFTER db_host,
    ADD COLUMN db_user VARCHAR(100) NOT NULL DEFAULT 'root' AFTER db_port,
    ADD COLUMN db_password VARCHAR(255) NOT NULL DEFAULT '' AFTER db_user,
    ADD COLUMN db_name VARCHAR(100) NOT NULL DEFAULT '' AFTER db_password;
