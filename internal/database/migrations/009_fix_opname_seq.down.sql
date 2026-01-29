-- Revert to original (flawed) function if needed
CREATE OR REPLACE FUNCTION generate_opname_session_code()
RETURNS TEXT AS $$
DECLARE
    today TEXT;
    seq INTEGER;
    code TEXT;
BEGIN
    today := TO_CHAR(NOW(), 'YYYYMMDD');
    
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(session_code FROM 12) AS INTEGER)
    ), 0) + 1 INTO seq
    FROM stock_opname_sessions
    WHERE session_code LIKE 'SO-' || today || '-%';
    
    code := 'SO-' || today || '-' || LPAD(seq::TEXT, 4, '0');
    RETURN code;
END;
$$ LANGUAGE plpgsql;
