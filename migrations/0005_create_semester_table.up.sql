-- Migration: Create semester table (PostgreSQL-compatible)
-- NOTE: Converted from MySQL to PostgreSQL (removed ENGINE, ENUM -> CHECK)

CREATE TABLE IF NOT EXISTS semester (
  id_semester CHAR(6) PRIMARY KEY,
  tahun_ajaran VARCHAR(9) NOT NULL,
  term TEXT NOT NULL CHECK (term IN ('Ganjil','Genap','Antara')),
  tanggal_mulai DATE,
  tanggal_selesai DATE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Function to auto-update updated_at on row updates (scoped to semester)
CREATE OR REPLACE FUNCTION trigger_set_timestamp_semester()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Ensure old trigger is removed if exists, then create a fresh one
DROP TRIGGER IF EXISTS set_timestamp ON semester;
CREATE TRIGGER set_timestamp
BEFORE UPDATE ON semester
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp_semester();