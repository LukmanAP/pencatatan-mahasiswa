-- Migration: Create prodi table (PostgreSQL-compatible)
-- NOTE: Converted from MySQL to PostgreSQL (removed ENGINE, used CHECK constraints for ENUM, replaced ON UPDATE with trigger)

CREATE TABLE IF NOT EXISTS prodi (
  id_prodi CHAR(8) PRIMARY KEY,
  id_fakultas CHAR(8) NOT NULL,
  nama_prodi VARCHAR(120) NOT NULL,
  jenjang TEXT NOT NULL CHECK (jenjang IN ('D3','D4','S1','S2','S3')),
  akreditasi TEXT NULL CHECK (akreditasi IN ('A','B','C','Baik','Baik Sekali','Unggul')),
  kode_prodi VARCHAR(16) NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_prodi_fakultas FOREIGN KEY (id_fakultas) REFERENCES fakultas(id_fakultas)
    ON UPDATE CASCADE ON DELETE RESTRICT
);

-- Function to auto-update updated_at on row updates (scoped to prodi)
CREATE OR REPLACE FUNCTION trigger_set_timestamp_prodi()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Ensure old trigger is removed if exists, then create a fresh one
DROP TRIGGER IF EXISTS set_timestamp ON prodi;
CREATE TRIGGER set_timestamp
BEFORE UPDATE ON prodi
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp_prodi();