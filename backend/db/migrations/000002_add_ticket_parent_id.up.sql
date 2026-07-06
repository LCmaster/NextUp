ALTER TABLE tickets ADD COLUMN parent_id UUID REFERENCES tickets(id) ON DELETE CASCADE;
CREATE INDEX idx_tickets_parent_id ON tickets(parent_id);
