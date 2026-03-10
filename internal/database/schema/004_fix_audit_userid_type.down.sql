-- Reverse: convert user_id back to VARCHAR
ALTER TABLE audit_logs ALTER COLUMN user_id TYPE VARCHAR(255) USING user_id::text;

-- Restore NOT NULL with 'anonymous' as default for NULLs
UPDATE audit_logs SET user_id = 'anonymous' WHERE user_id IS NULL;
ALTER TABLE audit_logs ALTER COLUMN user_id SET NOT NULL;
