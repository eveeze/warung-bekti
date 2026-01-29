-- Migration: 007_consignment (DOWN)
DROP FUNCTION IF EXISTS generate_settlement_number();
DROP TABLE IF EXISTS consignment_settlement_items;
DROP TABLE IF EXISTS consignment_settlements;
ALTER TABLE products DROP COLUMN IF EXISTS consignor_id;
ALTER TABLE products DROP COLUMN IF EXISTS commission_rate;
ALTER TABLE products DROP COLUMN IF EXISTS is_consignment;
DROP TABLE IF EXISTS consignors;
DROP TYPE IF EXISTS settlement_status;
