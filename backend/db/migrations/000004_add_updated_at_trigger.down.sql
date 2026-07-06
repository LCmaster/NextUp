-- Remove triggers first, then the shared function.
DROP TRIGGER IF EXISTS set_updated_at_tickets ON tickets;
DROP TRIGGER IF EXISTS set_updated_at_projects ON projects;
DROP TRIGGER IF EXISTS set_updated_at_users ON users;

DROP FUNCTION IF EXISTS trigger_set_updated_at();
