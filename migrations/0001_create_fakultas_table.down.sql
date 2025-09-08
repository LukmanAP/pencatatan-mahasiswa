-- Rollback migration: Drop trigger, function and table for fakultas

-- Drop trigger if exists
DROP TRIGGER IF EXISTS set_timestamp ON fakultas;

-- Drop function if exists
DROP FUNCTION IF EXISTS trigger_set_timestamp_fakultas();

-- Drop table
DROP TABLE IF EXISTS fakultas;