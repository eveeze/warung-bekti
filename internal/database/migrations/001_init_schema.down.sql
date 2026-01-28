-- =============================================
-- Migration: 001_init_schema (DOWN)
-- Description: Rollback initial database schema
-- =============================================

-- Drop triggers
DROP TRIGGER IF EXISTS update_app_settings_updated_at ON app_settings;
DROP TRIGGER IF EXISTS update_daily_summaries_updated_at ON daily_summaries;
DROP TRIGGER IF EXISTS update_purchases_updated_at ON purchases;
DROP TRIGGER IF EXISTS update_suppliers_updated_at ON suppliers;
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;
DROP TRIGGER IF EXISTS update_customers_updated_at ON customers;
DROP TRIGGER IF EXISTS update_pricing_tiers_updated_at ON pricing_tiers;
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;

-- Drop functions
DROP FUNCTION IF EXISTS generate_purchase_number();
DROP FUNCTION IF EXISTS generate_invoice_number();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (in reverse order of creation due to foreign keys)
DROP TABLE IF EXISTS app_settings;
DROP TABLE IF EXISTS daily_summaries;
DROP TABLE IF EXISTS purchase_items;
DROP TABLE IF EXISTS purchases;
DROP TABLE IF EXISTS suppliers;
DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS kasbon_records;
DROP TABLE IF EXISTS transaction_items;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS pricing_tiers;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;

-- Drop types
DROP TYPE IF EXISTS purchase_status;
DROP TYPE IF EXISTS stock_movement_type;
DROP TYPE IF EXISTS kasbon_type;
DROP TYPE IF EXISTS transaction_status;
DROP TYPE IF EXISTS payment_method;
