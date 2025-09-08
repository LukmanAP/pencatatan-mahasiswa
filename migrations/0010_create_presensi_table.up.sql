-- Migration: Create presensi table (PostgreSQL-compatible)
-- NOTE: Converted from MySQL to PostgreSQL (AUTO_INCREMENT -> IDENTITY, ENUM -> CHECK)

CREATE TABLE IF NOT EXISTS presensi (
  id_presensi BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  id_kelas CHAR(12) NOT NULL,
  pertemuan_ke SMALLINT NOT NULL,
  tanggal DATE,
  id_mahasiswa CHAR(12) NOT NULL,
  status_hadir TEXT DEFAULT 'Hadir' CHECK (status_hadir IN ('Hadir','Sakit','Izin','Alpa')),
  CONSTRAINT fk_presensi_kelas FOREIGN KEY (id_kelas) REFERENCES kelas_kuliah(id_kelas)
    ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT fk_presensi_mhs FOREIGN KEY (id_mahasiswa) REFERENCES mahasiswa(id_mahasiswa)
    ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT uq_presensi UNIQUE (id_kelas, pertemuan_ke, id_mahasiswa)
);