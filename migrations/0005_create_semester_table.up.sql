-- Migration: Create semester table (PostgreSQL-compatible)
-- NOTE: Converted from MySQL to PostgreSQL (removed ENGINE, ENUM -> CHECK)

CREATE TABLE IF NOT EXISTS semester (
  id_semester CHAR(6) PRIMARY KEY,
  tahun_ajaran VARCHAR(9) NOT NULL,
  term TEXT NOT NULL CHECK (term IN ('Ganjil','Genap','Antara')),
  tanggal_mulai DATE,
  tanggal_selesai DATE
);