-- Migration: Create nilai table (PostgreSQL-compatible)
-- NOTE: Converted from MySQL to PostgreSQL (ENUM -> CHECK, ON UPDATE -> trigger)

CREATE TABLE IF NOT EXISTS nilai (
  id_krs BIGINT PRIMARY KEY,
  nilai_angka NUMERIC(5,2),
  nilai_huruf TEXT CHECK (nilai_huruf IN ('A','AB','B','BC','C','D','E')),
  bobot NUMERIC(4,2),
  tgl_input TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  tgl_update TIMESTAMP NULL,
  CONSTRAINT fk_nilai_krs FOREIGN KEY (id_krs) REFERENCES krs(id_krs)
    ON UPDATE CASCADE ON DELETE CASCADE
);

-- Trigger to set tgl_update on row update
CREATE OR REPLACE FUNCTION trigger_set_tgl_update_nilai()
RETURNS TRIGGER AS $$
BEGIN
  NEW.tgl_update = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS set_tgl_update ON nilai;
CREATE TRIGGER set_tgl_update
BEFORE UPDATE ON nilai
FOR EACH ROW
EXECUTE FUNCTION trigger_set_tgl_update_nilai();