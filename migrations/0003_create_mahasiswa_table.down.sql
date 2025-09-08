-- Rollback migration: Drop triggers, functions and table for mahasiswa

-- Drop triggers if exists
DROP TRIGGER IF EXISTS set_timestamp ON mahasiswa;
DROP TRIGGER IF EXISTS set_angkatan ON mahasiswa;

-- Drop functions if exists
DROP FUNCTION IF EXISTS trigger_set_timestamp_mahasiswa();
DROP FUNCTION IF EXISTS trigger_set_angkatan_mahasiswa();

-- Drop table
DROP TABLE IF EXISTS mahasiswa;