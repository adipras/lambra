-- Remove database configuration columns from projects table
ALTER TABLE projects
    DROP COLUMN db_host,
    DROP COLUMN db_port,
    DROP COLUMN db_user,
    DROP COLUMN db_password,
    DROP COLUMN db_name;
