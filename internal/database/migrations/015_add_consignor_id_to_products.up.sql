ALTER TABLE products ADD COLUMN consignor_id UUID REFERENCES consignors(id);
CREATE INDEX idx_products_consignor_id ON products(consignor_id);
