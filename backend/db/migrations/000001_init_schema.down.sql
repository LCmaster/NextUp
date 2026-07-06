DROP INDEX IF EXISTS idx_todos_project_id;
DROP INDEX IF EXISTS idx_tickets_status;
DROP INDEX IF EXISTS idx_tickets_assignee_id;
DROP INDEX IF EXISTS idx_tickets_project_id;
DROP INDEX IF EXISTS idx_projects_owner_id;

DROP TABLE IF EXISTS todos;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;
