-- Drop indexes first
DROP INDEX IF EXISTS idx_audit_logs_user_id;
DROP INDEX IF EXISTS idx_audit_logs_resource;
DROP INDEX IF EXISTS idx_audit_logs_timestamp;

-- Drop the audit_logs table
DROP TABLE IF EXISTS audit_logs;
