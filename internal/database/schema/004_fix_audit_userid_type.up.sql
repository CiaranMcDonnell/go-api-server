-- Convert user_id from VARCHAR to INTEGER for efficient JOINs with users table.
-- "anonymous" entries become NULL.

-- First, convert "anonymous" to NULL
UPDATE audit_logs SET user_id = NULL WHERE user_id = 'anonymous' OR user_id = '';

-- Alter column type
ALTER TABLE audit_logs ALTER COLUMN user_id TYPE INTEGER USING user_id::integer;

-- Allow NULL for unauthenticated requests
ALTER TABLE audit_logs ALTER COLUMN user_id DROP NOT NULL;
