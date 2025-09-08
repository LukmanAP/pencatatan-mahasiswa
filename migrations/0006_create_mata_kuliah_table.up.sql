-- Migration: Create mata_kuliah table (PostgreSQL-compatible)
-- NOTE: Converted from MySQL to PostgreSQL (ENGINE removed, ON UPDATE replaced by trigger)

CREATE TABLE IF NOT EXISTS mata_kuliah (
  id_mk CHAR(10) PRIMARY KEY,
  kode_mk VARCHAR(16) NOT NULL UNIQUE,
  nama_mk VARCHAR(120) NOT NULL,
  sks SMALLINT NOT NULL CHECK (sks >= 0),
  id_prodi CHAR(8) NOT NULL,
  id_dosen_pj CHAR(10),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_mk_prodi FOREIGN KEY (id_prodi) REFERENCES prodi(id_prodi)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT fk_mk_dosenpj FOREIGN KEY (id_dosen_pj) REFERENCES dosen(id_dosen)
    ON UPDATE CASCADE ON DELETE SET NULL
);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION trigger_set_timestamp_mk()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS set_timestamp ON mata_kuliah;
CREATE TRIGGER set_timestamp
BEFORE UPDATE ON mata_kuliah
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp_mk();