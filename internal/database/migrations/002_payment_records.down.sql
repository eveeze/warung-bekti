-- =============================================
-- Migration: 002_payment_records (DOWN)
-- Description: Rollback payment gateway tables
-- =============================================

ALTER TABLE transactions DROP COLUMN IF EXISTS payment_record_id;

DROP TABLE IF EXISTS payment_records;

DROP TYPE IF EXISTS payment_status;
