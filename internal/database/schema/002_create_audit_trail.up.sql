-- Create audit_trail table for tracking system activities
CREATE TABLE
    IF NOT EXISTS audit_logs (
        id SERIAL PRIMARY KEY,
        user_id VARCHAR(255) NOT NULL,
        attempted_identifier VARCHAR(255) NULL, -- Identifier used in login attempt
        action VARCHAR(255) NOT NULL,
        resource VARCHAR(255) NOT NULL,
        entity_id VARCHAR(255) NULL, -- Generic reference to a domain entity ID
        entity_type VARCHAR(255) NULL, -- Type of the referenced entity (e.g., 'user', 'order')
        request_path VARCHAR(255) NOT NULL,
        method VARCHAR(50) NOT NULL,
        status_code INT NOT NULL,
        ip_address VARCHAR(50) NOT NULL,
        user_agent TEXT,
        request_body TEXT,
        timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

-- Add index for faster querying on common fields
CREATE INDEX idx_audit_logs_user_id ON audit_logs (user_id);

CREATE INDEX idx_audit_logs_resource ON audit_logs (resource);

CREATE INDEX idx_audit_logs_timestamp ON audit_logs (timestamp);

COMMENT ON TABLE audit_logs IS 'System audit trail for tracking user actions and API requests';

COMMENT ON COLUMN audit_logs.attempted_identifier IS 'Identifier used in login attempt, if applicable (e.g., username or email).';

COMMENT ON COLUMN audit_logs.entity_id IS 'Generic reference to a domain entity ID, if applicable.';

COMMENT ON COLUMN audit_logs.entity_type IS 'Type of the referenced entity (e.g., user, order), if applicable.';
