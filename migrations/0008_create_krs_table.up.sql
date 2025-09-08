-- Migration: Create krs table (PostgreSQL-compatible)
-- NOTE: Converted from MySQL to PostgreSQL (AUTO_INCREMENT -> GENERATED ALWAYS AS IDENTITY, ENUM -> CHECK)

CREATE TABLE IF NOT EXISTS krs (
  id_krs BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  id_mahasiswa CHAR(12) NOT NULL,
  id_kelas CHAR(12) NOT NULL,
  id_semester CHAR(6) NOT NULL,
  tanggal_daftar TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status_krs TEXT DEFAULT 'Diambil' CHECK (status_krs IN ('Diambil','Batal')),
  CONSTRAINT uq_krs UNIQUE (id_mahasiswa, id_kelas),
  CONSTRAINT fk_krs_mhs FOREIGN KEY (id_mahasiswa) REFERENCES mahasiswa(id_mahasiswa)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT fk_krs_kelas FOREIGN KEY (id_kelas) REFERENCES kelas_kuliah(id_kelas)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT fk_krs_sem FOREIGN KEY (id_semester) REFERENCES semester(id_semester)
    ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_krs_mhs_sem ON krs(id_mahasiswa, id_semester);