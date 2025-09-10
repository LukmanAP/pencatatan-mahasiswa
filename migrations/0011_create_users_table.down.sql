-- Rollback migration: Drop users and trigger

DROP TRIGGER IF EXISTS set_timestamp ON users;
DROP FUNCTION IF EXISTS trigger_set_timestamp_users();
DROP TABLE IF EXISTS users;