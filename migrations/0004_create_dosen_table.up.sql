-- Migration: Create dosen table (PostgreSQL-compatible)
-- NOTE: Converted from MySQL to PostgreSQL (removed ON UPDATE in column, replaced with trigger)

CREATE TABLE IF NOT EXISTS dosen (
  id_dosen CHAR(10) PRIMARY KEY,
  nidn VARCHAR(16) UNIQUE,
  nama_dosen VARCHAR(120) NOT NULL,
  email VARCHAR(120),
  no_hp VARCHAR(20),
  jabatan_akademik VARCHAR(60),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Function to auto-update updated_at on row updates (scoped to dosen)
CREATE OR REPLACE FUNCTION trigger_set_timestamp_dosen()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Ensure old trigger is removed if exists, then create a fresh one
DROP TRIGGER IF EXISTS set_timestamp ON dosen;
CREATE TRIGGER set_timestamp
BEFORE UPDATE ON dosen
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp_dosen();