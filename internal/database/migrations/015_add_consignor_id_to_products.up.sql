ALTER TABLE products ADD COLUMN IF NOT EXISTS consignor_id UUID REFERENCES consignors(id);
CREATE INDEX IF NOT EXISTS idx_products_consignor_id ON products(consignor_id);
