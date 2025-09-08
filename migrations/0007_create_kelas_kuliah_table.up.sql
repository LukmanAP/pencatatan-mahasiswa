-- Migration: Create kelas_kuliah table (PostgreSQL-compatible)
-- NOTE: Converted from MySQL to PostgreSQL (ENGINE removed, ENUM -> CHECK)

CREATE TABLE IF NOT EXISTS kelas_kuliah (
  id_kelas CHAR(12) PRIMARY KEY,
  id_mk CHAR(10) NOT NULL,
  id_semester CHAR(6) NOT NULL,
  nama_kelas VARCHAR(10) NOT NULL,
  id_dosen_pengampu CHAR(10) NOT NULL,
  kapasitas_max INT DEFAULT 40,
  jadwal_hari TEXT CHECK (jadwal_hari IN ('Senin','Selasa','Rabu','Kamis','Jumat','Sabtu','Minggu')),
  jadwal_mulai TIME,
  jadwal_selesai TIME,
  ruangan VARCHAR(30),
  CONSTRAINT fk_kelas_mk FOREIGN KEY (id_mk) REFERENCES mata_kuliah(id_mk)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT fk_kelas_semester FOREIGN KEY (id_semester) REFERENCES semester(id_semester)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT fk_kelas_dosen FOREIGN KEY (id_dosen_pengampu) REFERENCES dosen(id_dosen)
    ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_kelas_mk_sem ON kelas_kuliah(id_mk, id_semester);