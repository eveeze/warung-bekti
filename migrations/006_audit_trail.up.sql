-- =============================================
-- Migration: 006_audit_trail
-- Description: Audit logging system
-- =============================================

-- =============================================
-- Audit Action Enum
-- =============================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'audit_action') THEN
        CREATE TYPE audit_action AS ENUM (
            'create', 'update', 'delete', 'login', 'logout',
            'view', 'export', 'import', 'approve', 'reject'
        );
    END IF;
END
$$;

-- =============================================
-- Audit Logs Table
-- =============================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,                             -- Can be null for system actions
    username VARCHAR(100),                    -- Snapshot of username
    user_role VARCHAR(50),                    -- Snapshot of role
    action audit_action NOT NULL,
    entity_type VARCHAR(100) NOT NULL,        -- 'product', 'transaction', 'customer', etc.
    entity_id UUID,
    entity_name VARCHAR(255),                 -- Quick reference (e.g., product name)
    old_values JSONB,                         -- Previous state
    new_values JSONB,                         -- New state
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_id VARCHAR(100),                  -- For tracing
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_date_range ON audit_logs(created_at);

-- Partial index for important actions
CREATE INDEX idx_audit_logs_important ON audit_logs(created_at DESC)
    WHERE action IN ('delete', 'approve', 'reject');
