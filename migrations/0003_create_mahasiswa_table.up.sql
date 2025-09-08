-- Migration: Create mahasiswa table (PostgreSQL-compatible)
-- NOTE: Converted from MySQL to PostgreSQL (removed ENGINE, used CHECK constraints for ENUM, replaced ON UPDATE with triggers)

CREATE TABLE IF NOT EXISTS mahasiswa (
  id_mahasiswa CHAR(12) PRIMARY KEY,
  id_prodi CHAR(8) NOT NULL,
  nik CHAR(16) UNIQUE,
  nama_lengkap VARCHAR(120) NOT NULL,
  jenis_kelamin TEXT NOT NULL CHECK (jenis_kelamin IN ('L','P')),
  tempat_lahir VARCHAR(80),
  tanggal_lahir DATE,
  alamat TEXT,
  email VARCHAR(120),
  no_hp VARCHAR(20),
  tahun_masuk SMALLINT NOT NULL CHECK (tahun_masuk >= 1900),
  status TEXT DEFAULT 'Aktif' CHECK (status IN ('Aktif','Cuti','Lulus','Drop Out','Non-Aktif')),
  angkatan SMALLINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_mhs_prodi FOREIGN KEY (id_prodi) REFERENCES prodi(id_prodi)
    ON UPDATE CASCADE ON DELETE RESTRICT
);

-- Function to auto-update updated_at on row updates (scoped to mahasiswa)
CREATE OR REPLACE FUNCTION trigger_set_timestamp_mahasiswa()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Ensure old trigger is removed if exists, then create a fresh one for timestamp
DROP TRIGGER IF EXISTS set_timestamp ON mahasiswa;
CREATE TRIGGER set_timestamp
BEFORE UPDATE ON mahasiswa
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp_mahasiswa();

-- Function to keep angkatan equal to tahun_masuk
CREATE OR REPLACE FUNCTION trigger_set_angkatan_mahasiswa()
RETURNS TRIGGER AS $$
BEGIN
  NEW.angkatan = NEW.tahun_masuk;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Ensure old trigger is removed if exists, then create trigger for angkatan
DROP TRIGGER IF EXISTS set_angkatan ON mahasiswa;
CREATE TRIGGER set_angkatan
BEFORE INSERT OR UPDATE ON mahasiswa
FOR EACH ROW
EXECUTE FUNCTION trigger_set_angkatan_mahasiswa();