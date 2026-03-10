CREATE INDEX idx_items_user_id_created_at ON items (user_id, created_at DESC);
DROP INDEX IF EXISTS idx_items_user_id;
DROP INDEX IF EXISTS idx_items_created_at;
