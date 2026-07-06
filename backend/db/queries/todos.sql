-- name: CreateTodo :one
INSERT INTO todos (project_id, title)
VALUES ($1, $2)
RETURNING *;

-- name: GetTodoByID :one
SELECT * FROM todos WHERE id = $1;

-- name: ListTodosByProject :many
SELECT * FROM todos WHERE project_id = $1 ORDER BY created_at DESC;

-- name: UpdateTodo :one
UPDATE todos
SET title = $2,
    is_completed = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTodo :exec
DELETE FROM todos WHERE id = $1;
