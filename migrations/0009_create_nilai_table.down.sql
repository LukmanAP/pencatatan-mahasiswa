-- Rollback migration: Drop nilai and trigger

DROP TRIGGER IF EXISTS set_tgl_update ON nilai;
DROP FUNCTION IF EXISTS trigger_set_tgl_update_nilai();
DROP TABLE IF EXISTS nilai;