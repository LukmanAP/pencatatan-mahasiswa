-- Rollback migration: Drop mata_kuliah table and trigger

DROP TRIGGER IF EXISTS set_timestamp ON mata_kuliah;
DROP FUNCTION IF EXISTS trigger_set_timestamp_mk();
DROP TABLE IF EXISTS mata_kuliah;