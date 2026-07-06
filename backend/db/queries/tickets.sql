-- name: CreateTicket :one
INSERT INTO tickets (project_id, title, description, status, priority, assignee_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetTicketByID :one
SELECT * FROM tickets WHERE id = $1;

-- name: ListTicketsByProject :many
SELECT * FROM tickets WHERE project_id = $1 ORDER BY created_at DESC;

-- name: UpdateTicket :one
UPDATE tickets
SET title = $2,
    description = $3,
    status = $4,
    priority = $5,
    assignee_id = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTicket :exec
DELETE FROM tickets WHERE id = $1;
