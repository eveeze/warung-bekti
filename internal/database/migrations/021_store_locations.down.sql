-- =============================================
-- Migration: 012_store_locations (DOWN)
-- =============================================

ALTER TABLE products DROP COLUMN IF EXISTS location_id;

DROP TRIGGER IF EXISTS update_locations_updated_at ON locations;
DROP TABLE IF EXISTS locations;

DROP TYPE IF EXISTS location_type;
