-- =============================================
-- Migration: 003_stock_opname (DOWN)
-- Description: Rollback stock opname tables
-- =============================================

DROP FUNCTION IF EXISTS generate_opname_session_code();

DROP TABLE IF EXISTS shopping_list_items;
DROP TABLE IF EXISTS stock_opname_items;
DROP TABLE IF EXISTS stock_opname_sessions;

ALTER TABLE stock_movements 
DROP COLUMN IF EXISTS expiry_date,
DROP COLUMN IF EXISTS batch_number;

DROP TYPE IF EXISTS opname_status;
