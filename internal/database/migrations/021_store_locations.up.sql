-- =============================================
-- Migration: 012_store_locations
-- Description: Adds locations table for 2D/3D store layouts and links products to locations
-- =============================================

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'location_type') THEN
        CREATE TYPE location_type AS ENUM ('shelf', 'fridge', 'showcase', 'floor', 'warehouse', 'cashier_area');
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    category location_type NOT NULL DEFAULT 'shelf',
    
    -- 3D Visual / Layout Coordinates
    x_coordinate DECIMAL(10,2) DEFAULT 0,
    y_coordinate DECIMAL(10,2) DEFAULT 0,
    z_coordinate DECIMAL(10,2) DEFAULT 0,
    
    -- Physical Dimensions
    width DECIMAL(10,2) DEFAULT 0,
    depth DECIMAL(10,2) DEFAULT 0,
    height DECIMAL(10,2) DEFAULT 0,
    
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Apply updated_at trigger for locations
CREATE TRIGGER update_locations_updated_at BEFORE UPDATE ON locations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add relation to products
ALTER TABLE products ADD COLUMN location_id UUID REFERENCES locations(id) ON DELETE SET NULL;
CREATE INDEX idx_products_location ON products(location_id);
