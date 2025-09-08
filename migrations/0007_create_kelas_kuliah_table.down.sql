-- Rollback migration: Drop kelas_kuliah and index

DROP INDEX IF EXISTS idx_kelas_mk_sem;
DROP TABLE IF EXISTS kelas_kuliah;