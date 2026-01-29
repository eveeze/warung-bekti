-- Migration: 004_cash_flow (DOWN)
DROP TABLE IF EXISTS cash_flow_records;
DROP TABLE IF EXISTS cash_drawer_sessions;
DROP TABLE IF EXISTS cash_flow_categories;
DROP TYPE IF EXISTS drawer_session_status;
DROP TYPE IF EXISTS cash_flow_type;
