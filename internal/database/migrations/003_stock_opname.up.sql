-- =============================================
-- Migration: 003_stock_opname
-- Description: Stock taking/opname system tables
-- =============================================

-- =============================================
-- Stock Opname Status Enum
-- =============================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'opname_status') THEN
        CREATE TYPE opname_status AS ENUM ('draft', 'in_progress', 'completed', 'cancelled');
    END IF;
END
$$;

-- =============================================
-- Stock Opname Sessions Table
-- =============================================
CREATE TABLE IF NOT EXISTS stock_opname_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_code VARCHAR(50) UNIQUE NOT NULL,   -- SO-YYYYMMDD-XXXX
    status opname_status DEFAULT 'draft',
    notes TEXT,
    
    -- Summary (calculated after completion)
    total_products INTEGER DEFAULT 0,
    total_variance INTEGER DEFAULT 0,        -- Sum of absolute variance
    total_loss_value BIGINT DEFAULT 0,       -- Value of lost items
    total_gain_value BIGINT DEFAULT 0,       -- Value of found items
    
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_by VARCHAR(100),
    completed_by VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_stock_opname_sessions_status ON stock_opname_sessions(status);
CREATE INDEX idx_stock_opname_sessions_created_at ON stock_opname_sessions(created_at DESC);

-- =============================================
-- Stock Opname Items Table
-- =============================================
CREATE TABLE IF NOT EXISTS stock_opname_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES stock_opname_sessions(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    
    -- Stock counts
    system_stock INTEGER NOT NULL,           -- Stock according to system at time of count
    physical_stock INTEGER NOT NULL,         -- Actual physical count
    variance INTEGER GENERATED ALWAYS AS (physical_stock - system_stock) STORED,
    
    -- Value calculation
    cost_per_unit BIGINT NOT NULL,           -- Cost price at time of count
    variance_value BIGINT GENERATED ALWAYS AS ((physical_stock - system_stock) * cost_per_unit) STORED,
    
    notes TEXT,
    counted_by VARCHAR(100),
    counted_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(session_id, product_id)
);

CREATE INDEX idx_stock_opname_items_session ON stock_opname_items(session_id);
CREATE INDEX idx_stock_opname_items_product ON stock_opname_items(product_id);
CREATE INDEX idx_stock_opname_items_variance ON stock_opname_items(variance) WHERE variance != 0;

-- =============================================
-- Add expiry_date to stock_movements (optional field)
-- =============================================
ALTER TABLE stock_movements 
ADD COLUMN IF NOT EXISTS expiry_date DATE,
ADD COLUMN IF NOT EXISTS batch_number VARCHAR(50);

CREATE INDEX idx_stock_movements_expiry ON stock_movements(expiry_date) 
    WHERE expiry_date IS NOT NULL;

-- =============================================
-- Shopping List (Auto-generated restock plan)
-- =============================================
CREATE TABLE IF NOT EXISTS shopping_list_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    current_stock INTEGER NOT NULL,
    min_stock INTEGER NOT NULL,
    suggested_qty INTEGER NOT NULL,          -- max_stock - current_stock, or default value
    estimated_cost BIGINT,                   -- suggested_qty * cost_price
    is_purchased BOOLEAN DEFAULT false,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_shopping_list_product ON shopping_list_items(product_id);
CREATE INDEX idx_shopping_list_purchased ON shopping_list_items(is_purchased);

-- =============================================
-- Triggers
-- =============================================
CREATE TRIGGER update_stock_opname_sessions_updated_at 
    BEFORE UPDATE ON stock_opname_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_shopping_list_items_updated_at 
    BEFORE UPDATE ON shopping_list_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================
-- Generate Session Code Function
-- =============================================
CREATE OR REPLACE FUNCTION generate_opname_session_code()
RETURNS TEXT AS $$
DECLARE
    today TEXT;
    seq INTEGER;
    code TEXT;
BEGIN
    today := TO_CHAR(NOW(), 'YYYYMMDD');
    
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(session_code FROM 12) AS INTEGER)
    ), 0) + 1 INTO seq
    FROM stock_opname_sessions
    WHERE session_code LIKE 'SO-' || today || '-%';
    
    code := 'SO-' || today || '-' || LPAD(seq::TEXT, 4, '0');
    RETURN code;
END;
$$ LANGUAGE plpgsql;
