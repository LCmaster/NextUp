-- name: CreateTicket :one
INSERT INTO tickets (project_id, title, description, status, priority, assignee_id, parent_id, creator_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetTicketByID :one
SELECT * FROM tickets WHERE id = $1;

-- name: ListTicketsByProjectAndUser :many
SELECT t.* FROM tickets t
JOIN project_members pm ON t.project_id = pm.project_id
WHERE t.project_id = $1 AND pm.user_id = $2
  AND (
    pm.role IN ('admin', 'owner')
    OR t.assignee_id = $2
    OR (t.creator_id = $2 AND t.assignee_id IS NULL)
  )
ORDER BY t.created_at DESC;

-- name: UpdateTicket :one
UPDATE tickets
SET title = $2,
    description = $3,
    status = $4,
    priority = $5,
    assignee_id = $6,
    parent_id = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTicket :exec
DELETE FROM tickets WHERE id = $1;
