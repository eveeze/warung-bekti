-- =============================================
-- Migration: 002_payment_records
-- Description: Payment gateway integration tables
-- =============================================

-- =============================================
-- Payment Status Enum
-- =============================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_status') THEN
        CREATE TYPE payment_status AS ENUM (
            'pending',      -- Menunggu pembayaran
            'settlement',   -- Pembayaran berhasil
            'capture',      -- Pembayaran tertangkap (CC)
            'deny',         -- Ditolak
            'cancel',       -- Dibatalkan
            'expire',       -- Kadaluarsa
            'failure',      -- Gagal
            'refund',       -- Dikembalikan
            'partial_refund' -- Dikembalikan sebagian
        );
    END IF;
END
$$;

-- =============================================
-- Payment Records Table
-- =============================================
CREATE TABLE IF NOT EXISTS payment_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    
    -- Midtrans Data
    order_id VARCHAR(100) UNIQUE NOT NULL,  -- Format: TRX-{transaction_id}-{timestamp}
    snap_token VARCHAR(500),                 -- Snap token untuk redirect
    redirect_url VARCHAR(500),               -- URL redirect ke Midtrans
    
    -- Payment Details
    payment_type VARCHAR(50),                -- qris, bank_transfer, gopay, etc.
    gross_amount BIGINT NOT NULL,
    currency VARCHAR(10) DEFAULT 'IDR',
    
    -- Status
    status payment_status DEFAULT 'pending',
    fraud_status VARCHAR(50),                -- accept, deny, challenge
    
    -- Midtrans Response (JSON)
    midtrans_response JSONB,
    
    -- Timestamps
    paid_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_payment_records_transaction ON payment_records(transaction_id);
CREATE INDEX idx_payment_records_order_id ON payment_records(order_id);
CREATE INDEX idx_payment_records_status ON payment_records(status);
CREATE INDEX idx_payment_records_created_at ON payment_records(created_at DESC);

-- =============================================
-- Add payment_record_id to transactions (optional link)
-- =============================================
ALTER TABLE transactions 
ADD COLUMN IF NOT EXISTS payment_record_id UUID REFERENCES payment_records(id) ON DELETE SET NULL;

-- =============================================
-- Trigger for updated_at
-- =============================================
CREATE TRIGGER update_payment_records_updated_at 
    BEFORE UPDATE ON payment_records
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
