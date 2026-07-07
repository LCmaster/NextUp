-- name: CreateProject :one
INSERT INTO projects (name, description, owner_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetProjectByID :one
SELECT * FROM projects WHERE id = $1 AND deleted_at IS NULL;

-- name: ListProjectsByMember :many
SELECT p.* FROM projects p
JOIN project_members pm ON p.id = pm.project_id
WHERE pm.user_id = $1 AND p.deleted_at IS NULL
ORDER BY p.created_at DESC;
-- name: UpdateProject :one
UPDATE projects
SET name = $2,
    description = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteProject :exec
UPDATE projects SET deleted_at = NOW() WHERE id = $1;
