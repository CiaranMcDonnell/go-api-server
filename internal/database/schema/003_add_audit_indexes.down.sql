-- Remove performance-related indexes from the audit_logs table

-- Drop the index for entity_id
DROP INDEX IF EXISTS idx_audit_logs_entity_id;

-- Drop the composite index for user and timestamp
DROP INDEX IF EXISTS idx_audit_logs_user_time;

-- Drop the index for attempted_identifier
DROP INDEX IF EXISTS idx_audit_logs_attempted_id;
