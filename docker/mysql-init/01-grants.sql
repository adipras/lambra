-- Grant the lambra user privileges to create per-service databases.
-- The platform validates each service's DB config and runs
-- CREATE DATABASE IF NOT EXISTS, which requires CREATE on *.* in dev.
GRANT ALL PRIVILEGES ON *.* TO 'lambra'@'%' WITH GRANT OPTION;
FLUSH PRIVILEGES;
