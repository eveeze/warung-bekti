-- =============================================
-- Migration: 007_consignment
-- Description: Consignment (Titip Jual) system
-- =============================================

-- =============================================
-- Consignors Table (Penitip Barang)
-- =============================================
CREATE TABLE IF NOT EXISTS consignors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    address TEXT,
    bank_account VARCHAR(50),                 -- For settlement payment
    bank_name VARCHAR(100),
    notes TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_consignors_name ON consignors USING gin(to_tsvector('indonesian', name));
CREATE INDEX idx_consignors_active ON consignors(is_active) WHERE is_active = true;

-- =============================================
-- Add consignment fields to products
-- =============================================
ALTER TABLE products
ADD COLUMN IF NOT EXISTS consignor_id UUID REFERENCES consignors(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS commission_rate DECIMAL(5,2) DEFAULT 0,  -- Persentase komisi warung (0-100)
ADD COLUMN IF NOT EXISTS is_consignment BOOLEAN DEFAULT false;

CREATE INDEX idx_products_consignment ON products(consignor_id) WHERE is_consignment = true;

-- =============================================
-- Consignment Settlement Status
-- =============================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'settlement_status') THEN
        CREATE TYPE settlement_status AS ENUM ('draft', 'confirmed', 'paid');
    END IF;
END
$$;

-- =============================================
-- Consignment Settlements Table
-- =============================================
CREATE TABLE IF NOT EXISTS consignment_settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    settlement_number VARCHAR(50) UNIQUE NOT NULL,  -- SET-YYYYMMDD-XXXX
    consignor_id UUID NOT NULL REFERENCES consignors(id),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    total_sales BIGINT NOT NULL,              -- Total penjualan barang titipan
    commission_amount BIGINT NOT NULL,        -- Bagian warung
    consignor_amount BIGINT NOT NULL,         -- Bagian penitip
    status settlement_status DEFAULT 'draft',
    notes TEXT,
    created_by VARCHAR(100),
    paid_by VARCHAR(100),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_settlements_consignor ON consignment_settlements(consignor_id);
CREATE INDEX idx_settlements_status ON consignment_settlements(status);
CREATE INDEX idx_settlements_period ON consignment_settlements(period_start, period_end);

-- =============================================
-- Consignment Settlement Items
-- =============================================
CREATE TABLE IF NOT EXISTS consignment_settlement_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    settlement_id UUID NOT NULL REFERENCES consignment_settlements(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    product_name VARCHAR(255) NOT NULL,
    quantity_sold INTEGER NOT NULL,
    unit_price BIGINT NOT NULL,
    total_sales BIGINT NOT NULL,
    commission_rate DECIMAL(5,2) NOT NULL,
    commission_amount BIGINT NOT NULL,
    consignor_amount BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_settlement_items_settlement ON consignment_settlement_items(settlement_id);

-- =============================================
-- Triggers
-- =============================================
CREATE TRIGGER update_consignors_updated_at 
    BEFORE UPDATE ON consignors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_consignment_settlements_updated_at 
    BEFORE UPDATE ON consignment_settlements
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================
-- Generate Settlement Number Function
-- =============================================
CREATE OR REPLACE FUNCTION generate_settlement_number()
RETURNS TEXT AS $$
DECLARE
    today TEXT;
    seq INTEGER;
    num TEXT;
BEGIN
    today := TO_CHAR(NOW(), 'YYYYMMDD');
    
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(settlement_number FROM 13) AS INTEGER)
    ), 0) + 1 INTO seq
    FROM consignment_settlements
    WHERE settlement_number LIKE 'SET-' || today || '-%';
    
    num := 'SET-' || today || '-' || LPAD(seq::TEXT, 4, '0');
    RETURN num;
END;
$$ LANGUAGE plpgsql;
