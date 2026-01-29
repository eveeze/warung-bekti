-- Note: This might fail if duplicates exist
DELETE FROM cash_drawer_sessions a USING cash_drawer_sessions b WHERE a.id < b.id AND a.session_date = b.session_date;
ALTER TABLE cash_drawer_sessions ADD CONSTRAINT cash_drawer_sessions_session_date_key UNIQUE(session_date);
