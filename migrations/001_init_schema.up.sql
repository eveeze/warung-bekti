-- =============================================
-- Migration: 001_init_schema
-- Description: Initial database schema setup
-- =============================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================
-- Categories Table
-- =============================================
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_categories_parent ON categories(parent_id);
CREATE INDEX idx_categories_active ON categories(is_active) WHERE is_active = true;

-- =============================================
-- Products Table (with hybrid stock support)
-- =============================================
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    barcode VARCHAR(50) UNIQUE,
    sku VARCHAR(50) UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    unit VARCHAR(50) NOT NULL DEFAULT 'pcs',  -- pcs, kg, liter, pack, dus, etc.
    base_price BIGINT NOT NULL,                -- harga jual dasar (dalam rupiah)
    cost_price BIGINT NOT NULL,                -- harga beli/HPP
    is_stock_active BOOLEAN DEFAULT true,      -- hybrid stock toggle
    current_stock INTEGER DEFAULT 0,
    min_stock_alert INTEGER DEFAULT 0,         -- batas minimum untuk alert restock
    max_stock INTEGER,                         -- optional: batas maksimum stok
    image_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT positive_base_price CHECK (base_price >= 0),
    CONSTRAINT positive_cost_price CHECK (cost_price >= 0),
    CONSTRAINT non_negative_stock CHECK (current_stock >= 0)
);

CREATE INDEX idx_products_barcode ON products(barcode) WHERE barcode IS NOT NULL;
CREATE INDEX idx_products_name ON products USING gin(to_tsvector('indonesian', name));
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_active ON products(is_active) WHERE is_active = true;
CREATE INDEX idx_products_low_stock ON products(current_stock, min_stock_alert) 
    WHERE is_stock_active = true AND is_active = true;

-- =============================================
-- Pricing Tiers (Variable Pricing Support)
-- =============================================
-- Untuk menangani harga bertingkat:
-- 1 pcs = 3500, 3+ pcs = 3000/pcs, 12+ pcs = 2800/pcs
CREATE TABLE IF NOT EXISTS pricing_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name VARCHAR(100),                         -- "Eceran", "Grosir", "Promo 3+", "Kartonan"
    min_quantity INTEGER NOT NULL DEFAULT 1,   -- minimal qty untuk tier ini
    max_quantity INTEGER,                      -- NULL = unlimited (sampai tier berikutnya)
    price BIGINT NOT NULL,                     -- harga per unit di tier ini
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_quantity_range CHECK (
        max_quantity IS NULL OR max_quantity >= min_quantity
    ),
    CONSTRAINT positive_tier_price CHECK (price >= 0),
    CONSTRAINT positive_min_quantity CHECK (min_quantity >= 1)
);

-- Composite index untuk efficient tier lookup
CREATE INDEX idx_pricing_tiers_product ON pricing_tiers(product_id);
CREATE INDEX idx_pricing_tiers_lookup ON pricing_tiers(product_id, min_quantity, is_active) 
    WHERE is_active = true;

-- =============================================
-- Customers (for Kasbon/Debt tracking)
-- =============================================
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    address TEXT,
    notes TEXT,
    credit_limit BIGINT DEFAULT 0,             -- batas maksimal kasbon
    current_debt BIGINT DEFAULT 0,             -- total hutang saat ini
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT non_negative_credit_limit CHECK (credit_limit >= 0),
    CONSTRAINT non_negative_debt CHECK (current_debt >= 0)
);

CREATE INDEX idx_customers_name ON customers USING gin(to_tsvector('indonesian', name));
CREATE INDEX idx_customers_phone ON customers(phone) WHERE phone IS NOT NULL;
CREATE INDEX idx_customers_with_debt ON customers(current_debt) WHERE current_debt > 0;

-- =============================================
-- Enum Types for Transactions
-- =============================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_method') THEN
        CREATE TYPE payment_method AS ENUM ('cash', 'kasbon', 'transfer', 'qris', 'mixed');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'transaction_status') THEN
        CREATE TYPE transaction_status AS ENUM ('pending', 'completed', 'cancelled', 'refunded');
    END IF;
END
$$;

-- =============================================
-- Transactions Table
-- =============================================
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    subtotal BIGINT NOT NULL,                  -- total sebelum diskon & pajak
    discount_amount BIGINT DEFAULT 0,          -- total diskon
    tax_amount BIGINT DEFAULT 0,               -- total pajak
    total_amount BIGINT NOT NULL,              -- total akhir
    payment_method payment_method NOT NULL DEFAULT 'cash',
    amount_paid BIGINT DEFAULT 0,              -- jumlah yang dibayar
    change_amount BIGINT DEFAULT 0,            -- kembalian
    status transaction_status DEFAULT 'completed',
    notes TEXT,
    cashier_name VARCHAR(100),                 -- nama kasir (simple, no user table for now)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT positive_subtotal CHECK (subtotal >= 0),
    CONSTRAINT non_negative_discount CHECK (discount_amount >= 0),
    CONSTRAINT non_negative_tax CHECK (tax_amount >= 0),
    CONSTRAINT positive_total CHECK (total_amount >= 0),
    CONSTRAINT non_negative_paid CHECK (amount_paid >= 0),
    CONSTRAINT non_negative_change CHECK (change_amount >= 0)
);

CREATE INDEX idx_transactions_invoice ON transactions(invoice_number);
CREATE INDEX idx_transactions_customer ON transactions(customer_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_payment ON transactions(payment_method);


-- =============================================
-- Transaction Items
-- =============================================
CREATE TABLE IF NOT EXISTS transaction_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    product_name VARCHAR(255) NOT NULL,        -- snapshot nama produk saat transaksi
    product_barcode VARCHAR(50),               -- snapshot barcode
    quantity INTEGER NOT NULL,
    unit VARCHAR(50) NOT NULL,                 -- snapshot unit
    unit_price BIGINT NOT NULL,                -- harga per unit saat transaksi
    cost_price BIGINT NOT NULL DEFAULT 0,      -- HPP untuk perhitungan profit
    subtotal BIGINT NOT NULL,                  -- quantity * unit_price
    discount_amount BIGINT DEFAULT 0,          -- diskon per item
    total_amount BIGINT NOT NULL,              -- subtotal - discount
    pricing_tier_id UUID REFERENCES pricing_tiers(id) ON DELETE SET NULL,
    pricing_tier_name VARCHAR(100),            -- snapshot tier name
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT positive_quantity CHECK (quantity > 0),
    CONSTRAINT positive_unit_price CHECK (unit_price >= 0),
    CONSTRAINT positive_item_subtotal CHECK (subtotal >= 0),
    CONSTRAINT positive_item_total CHECK (total_amount >= 0)
);

CREATE INDEX idx_transaction_items_transaction ON transaction_items(transaction_id);
CREATE INDEX idx_transaction_items_product ON transaction_items(product_id);

-- =============================================
-- Kasbon (Debt) Records
-- =============================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'kasbon_type') THEN
        CREATE TYPE kasbon_type AS ENUM ('debt', 'payment');
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS kasbon_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
    type kasbon_type NOT NULL,
    amount BIGINT NOT NULL,                    -- jumlah hutang atau pembayaran
    balance_before BIGINT NOT NULL,            -- saldo sebelum transaksi
    balance_after BIGINT NOT NULL,             -- saldo setelah transaksi
    notes TEXT,
    created_by VARCHAR(100),                   -- nama yang mencatat
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT positive_kasbon_amount CHECK (amount > 0),
    CONSTRAINT non_negative_balance_before CHECK (balance_before >= 0),
    CONSTRAINT non_negative_balance_after CHECK (balance_after >= 0)
);

CREATE INDEX idx_kasbon_customer ON kasbon_records(customer_id);
CREATE INDEX idx_kasbon_transaction ON kasbon_records(transaction_id);
CREATE INDEX idx_kasbon_created_at ON kasbon_records(created_at DESC);
CREATE INDEX idx_kasbon_type ON kasbon_records(type);

-- =============================================
-- Stock Movements (Audit Trail)
-- =============================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'stock_movement_type') THEN
        CREATE TYPE stock_movement_type AS ENUM (
            'initial',       -- stok awal
            'purchase',      -- pembelian/restock
            'sale',          -- penjualan
            'adjustment',    -- penyesuaian manual
            'return',        -- retur dari customer
            'damage',        -- barang rusak/expired
            'transfer_in',   -- transfer masuk (future)
            'transfer_out'   -- transfer keluar (future)
        );
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    type stock_movement_type NOT NULL,
    quantity INTEGER NOT NULL,                 -- positive for in, negative for out
    stock_before INTEGER NOT NULL,
    stock_after INTEGER NOT NULL,
    reference_type VARCHAR(50),                -- 'transaction', 'purchase', 'adjustment'
    reference_id UUID,                         -- ID dari transaksi/pembelian terkait
    cost_per_unit BIGINT,                      -- harga beli per unit (untuk restock)
    notes TEXT,
    created_by VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT non_negative_stock_before CHECK (stock_before >= 0),
    CONSTRAINT non_negative_stock_after CHECK (stock_after >= 0)
);

CREATE INDEX idx_stock_movements_product ON stock_movements(product_id);
CREATE INDEX idx_stock_movements_created_at ON stock_movements(created_at DESC);
CREATE INDEX idx_stock_movements_type ON stock_movements(type);
CREATE INDEX idx_stock_movements_reference ON stock_movements(reference_type, reference_id);

-- =============================================
-- Suppliers (for Purchase/Restock tracking)
-- =============================================
CREATE TABLE IF NOT EXISTS suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    address TEXT,
    notes TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_suppliers_name ON suppliers USING gin(to_tsvector('indonesian', name));

-- =============================================
-- Purchases (Restock Records)
-- =============================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'purchase_status') THEN
        CREATE TYPE purchase_status AS ENUM ('draft', 'ordered', 'received', 'cancelled');
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS purchases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_number VARCHAR(50) UNIQUE NOT NULL,
    supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL,
    total_amount BIGINT NOT NULL DEFAULT 0,
    status purchase_status DEFAULT 'received',
    notes TEXT,
    received_at TIMESTAMPTZ,
    created_by VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_purchases_supplier ON purchases(supplier_id);
CREATE INDEX idx_purchases_status ON purchases(status);
CREATE INDEX idx_purchases_created_at ON purchases(created_at DESC);

-- =============================================
-- Purchase Items
-- =============================================
CREATE TABLE IF NOT EXISTS purchase_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_id UUID NOT NULL REFERENCES purchases(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    cost_per_unit BIGINT NOT NULL,
    total_cost BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT positive_purchase_quantity CHECK (quantity > 0),
    CONSTRAINT positive_cost_per_unit CHECK (cost_per_unit >= 0)
);

CREATE INDEX idx_purchase_items_purchase ON purchase_items(purchase_id);
CREATE INDEX idx_purchase_items_product ON purchase_items(product_id);

-- =============================================
-- Daily Summary (for quick reporting)
-- =============================================
CREATE TABLE IF NOT EXISTS daily_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL UNIQUE,
    total_transactions INTEGER DEFAULT 0,
    total_sales BIGINT DEFAULT 0,              -- total penjualan
    total_cost BIGINT DEFAULT 0,               -- total HPP
    total_profit BIGINT DEFAULT 0,             -- estimasi profit
    total_kasbon_given BIGINT DEFAULT 0,       -- total kasbon yang diberikan
    total_kasbon_paid BIGINT DEFAULT 0,        -- total pembayaran kasbon
    cash_sales BIGINT DEFAULT 0,
    kasbon_sales BIGINT DEFAULT 0,
    transfer_sales BIGINT DEFAULT 0,
    qris_sales BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_daily_summaries_date ON daily_summaries(date DESC);

-- =============================================
-- App Settings (Key-Value store)
-- =============================================
CREATE TABLE IF NOT EXISTS app_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Insert default settings
INSERT INTO app_settings (key, value, description) VALUES
    ('store_name', 'Warung Kelontong', 'Nama toko'),
    ('store_address', '', 'Alamat toko'),
    ('store_phone', '', 'Nomor telepon toko'),
    ('tax_rate', '0', 'Persentase pajak (0-100)'),
    ('receipt_footer', 'Terima kasih telah berbelanja!', 'Footer struk'),
    ('invoice_prefix', 'INV', 'Prefix nomor invoice'),
    ('purchase_prefix', 'PO', 'Prefix nomor pembelian')
ON CONFLICT (key) DO NOTHING;

-- =============================================
-- Updated At Trigger Function
-- =============================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to all tables with updated_at
CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_pricing_tiers_updated_at BEFORE UPDATE ON pricing_tiers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_suppliers_updated_at BEFORE UPDATE ON suppliers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_purchases_updated_at BEFORE UPDATE ON purchases
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_daily_summaries_updated_at BEFORE UPDATE ON daily_summaries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_app_settings_updated_at BEFORE UPDATE ON app_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================
-- Invoice Number Generator Function
-- =============================================
CREATE OR REPLACE FUNCTION generate_invoice_number()
RETURNS TEXT AS $$
DECLARE
    prefix TEXT;
    today TEXT;
    seq INTEGER;
    invoice TEXT;
BEGIN
    SELECT value INTO prefix FROM app_settings WHERE key = 'invoice_prefix';
    IF prefix IS NULL THEN
        prefix := 'INV';
    END IF;
    
    today := TO_CHAR(NOW(), 'YYYYMMDD');
    
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(invoice_number FROM LENGTH(prefix) + 10) AS INTEGER)
    ), 0) + 1 INTO seq
    FROM transactions
    WHERE invoice_number LIKE prefix || today || '%';
    
    invoice := prefix || today || LPAD(seq::TEXT, 4, '0');
    RETURN invoice;
END;
$$ LANGUAGE plpgsql;

-- =============================================
-- Purchase Number Generator Function
-- =============================================
CREATE OR REPLACE FUNCTION generate_purchase_number()
RETURNS TEXT AS $$
DECLARE
    prefix TEXT;
    today TEXT;
    seq INTEGER;
    purchase_num TEXT;
BEGIN
    SELECT value INTO prefix FROM app_settings WHERE key = 'purchase_prefix';
    IF prefix IS NULL THEN
        prefix := 'PO';
    END IF;
    
    today := TO_CHAR(NOW(), 'YYYYMMDD');
    
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(purchase_number FROM LENGTH(prefix) + 10) AS INTEGER)
    ), 0) + 1 INTO seq
    FROM purchases
    WHERE purchase_number LIKE prefix || today || '%';
    
    purchase_num := prefix || today || LPAD(seq::TEXT, 4, '0');
    RETURN purchase_num;
END;
$$ LANGUAGE plpgsql;
