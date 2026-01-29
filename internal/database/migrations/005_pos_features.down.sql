-- Migration: 005_pos_features (DOWN)
DROP FUNCTION IF EXISTS generate_refund_number();
DROP FUNCTION IF EXISTS generate_hold_code();
DROP TABLE IF EXISTS refund_items;
DROP TABLE IF EXISTS refund_records;
DROP TABLE IF EXISTS held_cart_items;
DROP TABLE IF EXISTS held_carts;
DROP TYPE IF EXISTS refund_status;
DROP TYPE IF EXISTS held_cart_status;
