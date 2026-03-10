-- Index for entity_id to speed up entity-related queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_id ON audit_logs (entity_id);

-- Composite index for common user activity queries (user + time range)
-- Speeds up fetching a user's activity history within a specific period
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_time ON audit_logs (user_id, timestamp);

-- Index for attempted_identifier to speed up login attempt queries
-- Useful for security analysis and tracking failed login attempts
CREATE INDEX IF NOT EXISTS idx_audit_logs_attempted_id ON audit_logs (attempted_identifier);

COMMENT ON INDEX idx_audit_logs_entity_id IS 'Improves performance of queries filtering on entity_id.';

COMMENT ON INDEX idx_audit_logs_user_time IS 'Optimizes queries searching for user actions within specific time ranges.';

COMMENT ON INDEX idx_audit_logs_attempted_id IS 'Speeds up queries related to login attempts and security analysis.';
