-- Migration: 008_refillables (DOWN)
DROP TABLE IF EXISTS container_movements;
DROP TABLE IF EXISTS refillable_containers;
ALTER TABLE products DROP COLUMN IF EXISTS is_refillable;
ALTER TABLE products DROP COLUMN IF EXISTS empty_product_id;
ALTER TABLE products DROP COLUMN IF EXISTS full_product_id;
DROP TYPE IF EXISTS container_movement_type;
