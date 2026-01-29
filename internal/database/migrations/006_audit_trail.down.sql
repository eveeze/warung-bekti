-- Migration: 006_audit_trail (DOWN)
DROP TABLE IF EXISTS audit_logs;
DROP TYPE IF EXISTS audit_action;
