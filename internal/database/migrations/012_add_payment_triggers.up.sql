-- =============================================
-- Migration: 012_add_payment_triggers
-- Description: Add triggers to payment_records table (missed in 011)
-- =============================================

-- =============================================
-- Trigger for updated_at
-- =============================================
CREATE TRIGGER update_payment_records_updated_at 
    BEFORE UPDATE ON payment_records
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
