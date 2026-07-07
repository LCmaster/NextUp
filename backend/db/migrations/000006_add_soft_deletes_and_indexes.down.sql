DROP INDEX IF EXISTS idx_project_members_user_id;
DROP INDEX IF EXISTS idx_tickets_assignee_id;
DROP INDEX IF EXISTS idx_tickets_project_id;

ALTER TABLE tickets DROP COLUMN deleted_at;
ALTER TABLE projects DROP COLUMN deleted_at;
