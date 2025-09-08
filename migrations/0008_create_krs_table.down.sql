-- Rollback migration: Drop krs and related index

DROP INDEX IF EXISTS idx_krs_mhs_sem;
DROP TABLE IF EXISTS krs;