-- Rollback migration: Drop trigger, function and table for dosen

-- Drop trigger if exists
DROP TRIGGER IF EXISTS set_timestamp ON dosen;

-- Drop function if exists
DROP FUNCTION IF EXISTS trigger_set_timestamp_dosen();

-- Drop table
DROP TABLE IF EXISTS dosen;