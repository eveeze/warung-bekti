-- =============================================
-- Migration: 004_cash_flow
-- Description: Cash flow and drawer management tables
-- =============================================

-- =============================================
-- Cash Flow Type Enum
-- =============================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'cash_flow_type') THEN
        CREATE TYPE cash_flow_type AS ENUM ('income', 'expense');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'drawer_session_status') THEN
        CREATE TYPE drawer_session_status AS ENUM ('open', 'closed');
    END IF;
END
$$;

-- =============================================
-- Cash Flow Categories
-- =============================================
CREATE TABLE IF NOT EXISTS cash_flow_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    type cash_flow_type NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Insert default categories
INSERT INTO cash_flow_categories (name, type, description) VALUES
    ('Penjualan Tunai', 'income', 'Pendapatan dari penjualan cash'),
    ('Pembayaran Kasbon', 'income', 'Penerimaan pembayaran hutang'),
    ('Modal Tambahan', 'income', 'Penambahan modal ke laci'),
    ('Pembelian Stok', 'expense', 'Pembelian barang dagangan'),
    ('Listrik', 'expense', 'Bayar listrik'),
    ('Air', 'expense', 'Bayar air'),
    ('Sampah', 'expense', 'Iuran sampah'),
    ('Plastik/Kantong', 'expense', 'Beli plastik dan kantong'),
    ('Operasional Lain', 'expense', 'Pengeluaran operasional lainnya'),
    ('Pengambilan Pribadi', 'expense', 'Ambil uang untuk keperluan pribadi')
ON CONFLICT DO NOTHING;

-- =============================================
-- Cash Drawer Sessions Table
-- =============================================
CREATE TABLE IF NOT EXISTS cash_drawer_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_date DATE NOT NULL,
    opening_balance BIGINT NOT NULL,
    closing_balance BIGINT,
    expected_closing BIGINT,                -- Calculated expected amount
    difference BIGINT,                      -- closing - expected (+ = surplus, - = shortage)
    status drawer_session_status DEFAULT 'open',
    opened_by VARCHAR(100),
    closed_by VARCHAR(100),
    notes TEXT,
    opened_at TIMESTAMPTZ DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(session_date)
);

CREATE INDEX idx_drawer_sessions_date ON cash_drawer_sessions(session_date DESC);
CREATE INDEX idx_drawer_sessions_status ON cash_drawer_sessions(status);

-- =============================================
-- Cash Flow Records Table
-- =============================================
CREATE TABLE IF NOT EXISTS cash_flow_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    drawer_session_id UUID REFERENCES cash_drawer_sessions(id) ON DELETE SET NULL,
    category_id UUID REFERENCES cash_flow_categories(id) ON DELETE SET NULL,
    type cash_flow_type NOT NULL,
    amount BIGINT NOT NULL,
    description TEXT,
    reference_type VARCHAR(50),             -- 'transaction', 'kasbon_payment', 'manual'
    reference_id UUID,
    created_by VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT positive_amount CHECK (amount > 0)
);

CREATE INDEX idx_cash_flow_session ON cash_flow_records(drawer_session_id);
CREATE INDEX idx_cash_flow_date ON cash_flow_records(created_at DESC);
CREATE INDEX idx_cash_flow_type ON cash_flow_records(type);
CREATE INDEX idx_cash_flow_category ON cash_flow_records(category_id);

-- =============================================
-- Triggers
-- =============================================
CREATE TRIGGER update_cash_flow_categories_updated_at 
    BEFORE UPDATE ON cash_flow_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cash_drawer_sessions_updated_at 
    BEFORE UPDATE ON cash_drawer_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
