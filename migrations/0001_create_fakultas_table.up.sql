-- Migration: Create fakultas table (PostgreSQL-compatible)
-- NOTE: Converted from MySQL to PostgreSQL (removed ENGINE, replaced ON UPDATE with trigger)

CREATE TABLE IF NOT EXISTS fakultas (
  id_fakultas CHAR(8) PRIMARY KEY,
  nama_fakultas VARCHAR(100) NOT NULL,
  singkatan VARCHAR(20),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Function to auto-update updated_at on row updates (scoped to fakultas)
CREATE OR REPLACE FUNCTION trigger_set_timestamp_fakultas()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Ensure old trigger is removed if exists, then create a fresh one
DROP TRIGGER IF EXISTS set_timestamp ON fakultas;
CREATE TRIGGER set_timestamp
BEFORE UPDATE ON fakultas
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp_fakultas();