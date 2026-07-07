ALTER TABLE projects ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;
ALTER TABLE tickets ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;

CREATE INDEX idx_tickets_project_id ON tickets(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tickets_assignee_id ON tickets(assignee_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_project_members_user_id ON project_members(user_id);
