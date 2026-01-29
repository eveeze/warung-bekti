-- =============================================
-- Migration: 005_pos_features
-- Description: POS operational features - held carts and refunds
-- =============================================

-- =============================================
-- Held Cart Status Enum
-- =============================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'held_cart_status') THEN
        CREATE TYPE held_cart_status AS ENUM ('held', 'resumed', 'discarded');
    END IF;
END
$$;

-- =============================================
-- Held Carts Table
-- =============================================
CREATE TABLE IF NOT EXISTS held_carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hold_code VARCHAR(20) UNIQUE NOT NULL,   -- HOLD-001, HOLD-002, etc.
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    customer_name VARCHAR(255),               -- Quick reference
    status held_cart_status DEFAULT 'held',
    subtotal BIGINT DEFAULT 0,
    notes TEXT,
    held_by VARCHAR(100),
    resumed_by VARCHAR(100),
    held_at TIMESTAMPTZ DEFAULT NOW(),
    resumed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,                   -- Optional: auto-discard after X hours
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_held_carts_status ON held_carts(status);
CREATE INDEX idx_held_carts_held_at ON held_carts(held_at DESC);

-- =============================================
-- Held Cart Items Table
-- =============================================
CREATE TABLE IF NOT EXISTS held_cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES held_carts(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    product_name VARCHAR(255) NOT NULL,       -- Snapshot
    product_barcode VARCHAR(50),
    quantity INTEGER NOT NULL,
    unit VARCHAR(50) NOT NULL,
    unit_price BIGINT NOT NULL,
    subtotal BIGINT NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT positive_held_quantity CHECK (quantity > 0)
);

CREATE INDEX idx_held_cart_items_cart ON held_cart_items(cart_id);

-- =============================================
-- Refund Status Enum
-- =============================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'refund_status') THEN
        CREATE TYPE refund_status AS ENUM ('pending', 'approved', 'rejected', 'completed');
    END IF;
END
$$;

-- =============================================
-- Refund Records Table
-- =============================================
CREATE TABLE IF NOT EXISTS refund_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    refund_number VARCHAR(50) UNIQUE NOT NULL,  -- REF-YYYYMMDD-XXXX
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    total_refund_amount BIGINT NOT NULL,
    refund_method VARCHAR(20) NOT NULL,         -- 'cash', 'store_credit', 'original'
    status refund_status DEFAULT 'pending',
    reason TEXT NOT NULL,
    notes TEXT,
    requested_by VARCHAR(100),
    approved_by VARCHAR(100),
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_refund_records_transaction ON refund_records(transaction_id);
CREATE INDEX idx_refund_records_status ON refund_records(status);
CREATE INDEX idx_refund_records_date ON refund_records(created_at DESC);

-- =============================================
-- Refund Items Table
-- =============================================
CREATE TABLE IF NOT EXISTS refund_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    refund_id UUID NOT NULL REFERENCES refund_records(id) ON DELETE CASCADE,
    transaction_item_id UUID NOT NULL REFERENCES transaction_items(id),
    product_id UUID NOT NULL REFERENCES products(id),
    product_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price BIGINT NOT NULL,
    refund_amount BIGINT NOT NULL,
    reason TEXT,
    restock BOOLEAN DEFAULT true,              -- Whether to add back to inventory
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT positive_refund_quantity CHECK (quantity > 0)
);

CREATE INDEX idx_refund_items_refund ON refund_items(refund_id);

-- =============================================
-- Triggers
-- =============================================
CREATE TRIGGER update_held_carts_updated_at 
    BEFORE UPDATE ON held_carts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_refund_records_updated_at 
    BEFORE UPDATE ON refund_records
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================
-- Generate Hold Code Function
-- =============================================
CREATE OR REPLACE FUNCTION generate_hold_code()
RETURNS TEXT AS $$
DECLARE
    seq INTEGER;
    code TEXT;
BEGIN
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(hold_code FROM 6) AS INTEGER)
    ), 0) + 1 INTO seq
    FROM held_carts
    WHERE hold_code LIKE 'HOLD-%';
    
    code := 'HOLD-' || LPAD(seq::TEXT, 3, '0');
    RETURN code;
END;
$$ LANGUAGE plpgsql;

-- =============================================
-- Generate Refund Number Function
-- =============================================
CREATE OR REPLACE FUNCTION generate_refund_number()
RETURNS TEXT AS $$
DECLARE
    today TEXT;
    seq INTEGER;
    refund_num TEXT;
BEGIN
    today := TO_CHAR(NOW(), 'YYYYMMDD');
    
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(refund_number FROM 13) AS INTEGER)
    ), 0) + 1 INTO seq
    FROM refund_records
    WHERE refund_number LIKE 'REF-' || today || '-%';
    
    refund_num := 'REF-' || today || '-' || LPAD(seq::TEXT, 4, '0');
    RETURN refund_num;
END;
$$ LANGUAGE plpgsql;
