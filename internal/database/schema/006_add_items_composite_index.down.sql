DROP INDEX IF EXISTS idx_items_user_id_created_at;
CREATE INDEX idx_items_user_id ON items (user_id);
CREATE INDEX idx_items_created_at ON items (created_at);
