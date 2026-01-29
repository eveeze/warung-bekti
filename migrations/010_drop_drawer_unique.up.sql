ALTER TABLE cash_drawer_sessions DROP CONSTRAINT IF EXISTS cash_drawer_sessions_session_date_key;
CREATE INDEX IF NOT EXISTS idx_drawer_sessions_date ON cash_drawer_sessions(session_date);
