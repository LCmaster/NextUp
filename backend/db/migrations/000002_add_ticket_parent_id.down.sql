DROP INDEX IF EXISTS idx_tickets_parent_id;
ALTER TABLE tickets DROP COLUMN parent_id;
