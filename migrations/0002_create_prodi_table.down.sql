-- Rollback migration: Drop trigger, function and table for prodi

-- Drop trigger if exists
DROP TRIGGER IF EXISTS set_timestamp ON prodi;

-- Drop function if exists
DROP FUNCTION IF EXISTS trigger_set_timestamp_prodi();

-- Drop table
DROP TABLE IF EXISTS prodi;