-- Revert of 002 Logic to ensure table exists
-- If table was dropped or missing
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_status') THEN
        CREATE TYPE payment_status AS ENUM (
            'pending', 'settlement', 'capture', 'deny', 'cancel', 
            'expire', 'failure', 'refund', 'partial_refund'
        );
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS payment_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    order_id VARCHAR(100) UNIQUE NOT NULL,
    snap_token VARCHAR(500),
    redirect_url VARCHAR(500),
    payment_type VARCHAR(50),
    gross_amount BIGINT NOT NULL,
    currency VARCHAR(10) DEFAULT 'IDR',
    status payment_status DEFAULT 'pending',
    fraud_status VARCHAR(50),
    midtrans_response JSONB,
    paid_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_records_transaction ON payment_records(transaction_id);
CREATE INDEX IF NOT EXISTS idx_payment_records_order_id ON payment_records(order_id);
CREATE INDEX IF NOT EXISTS idx_payment_records_status ON payment_records(status);

-- Add column if missing (idempotent)
DO $$
BEGIN
    ALTER TABLE transactions ADD COLUMN payment_record_id UUID REFERENCES payment_records(id) ON DELETE SET NULL;
EXCEPTION
    WHEN duplicate_column THEN NULL;
END
$$;
