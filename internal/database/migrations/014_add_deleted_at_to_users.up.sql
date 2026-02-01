ALTER TABLE users ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
