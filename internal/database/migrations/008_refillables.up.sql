-- =============================================
-- Migration: 008_refillables
-- Description: Gas & Galon (Refillable Container) tracking
-- =============================================

-- =============================================
-- Add refillable fields to products
-- =============================================
ALTER TABLE products
ADD COLUMN IF NOT EXISTS is_refillable BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS empty_product_id UUID REFERENCES products(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS full_product_id UUID REFERENCES products(id) ON DELETE SET NULL;

-- empty_product_id: When selling "Gas Isi", this points to "Tabung Kosong" product
-- full_product_id: When selling "Tabung Kosong", this points to "Gas Isi" product

CREATE INDEX idx_products_refillable ON products(is_refillable) WHERE is_refillable = true;

-- =============================================
-- Refillable Container Inventory
-- Tracks container balance (kosong vs isi)
-- =============================================
CREATE TABLE IF NOT EXISTS refillable_containers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    container_type VARCHAR(50) NOT NULL,       -- 'gas_3kg', 'gas_12kg', 'galon_aqua', etc.
    empty_count INTEGER DEFAULT 0,
    full_count INTEGER DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(product_id),
    CONSTRAINT non_negative_empty CHECK (empty_count >= 0),
    CONSTRAINT non_negative_full CHECK (full_count >= 0)
);

CREATE INDEX idx_refillable_containers_product ON refillable_containers(product_id);
CREATE INDEX idx_refillable_containers_type ON refillable_containers(container_type);

-- =============================================
-- Container Movement Type
-- =============================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'container_movement_type') THEN
        CREATE TYPE container_movement_type AS ENUM (
            'sale_exchange',      -- Customer brings empty, gets full
            'restock_exchange',   -- Warung gives empty to agent, gets full
            'purchase_empty',     -- Buy empty containers
            'purchase_full',      -- Buy full containers
            'return_empty',       -- Customer returns empty deposit
            'adjustment'          -- Manual adjustment
        );
    END IF;
END
$$;

-- =============================================
-- Container Movements Table
-- =============================================
CREATE TABLE IF NOT EXISTS container_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    container_id UUID NOT NULL REFERENCES refillable_containers(id) ON DELETE CASCADE,
    type container_movement_type NOT NULL,
    empty_change INTEGER NOT NULL DEFAULT 0,  -- + for increase, - for decrease
    full_change INTEGER NOT NULL DEFAULT 0,
    empty_before INTEGER NOT NULL,
    empty_after INTEGER NOT NULL,
    full_before INTEGER NOT NULL,
    full_after INTEGER NOT NULL,
    reference_type VARCHAR(50),               -- 'transaction', 'purchase', 'manual'
    reference_id UUID,
    notes TEXT,
    created_by VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_container_movements_container ON container_movements(container_id);
CREATE INDEX idx_container_movements_date ON container_movements(created_at DESC);
CREATE INDEX idx_container_movements_type ON container_movements(type);

-- =============================================
-- Trigger
-- =============================================
CREATE TRIGGER update_refillable_containers_updated_at 
    BEFORE UPDATE ON refillable_containers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
