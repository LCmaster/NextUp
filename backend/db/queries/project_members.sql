-- name: AddProjectMember :one
INSERT INTO project_members (project_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (project_id, user_id) DO UPDATE SET role = EXCLUDED.role
RETURNING *;

-- name: GetProjectMember :one
SELECT * FROM project_members
WHERE project_id = $1 AND user_id = $2;

-- name: ListProjectMembers :many
SELECT pm.project_id, pm.user_id, pm.role, pm.created_at,
       u.first_name, u.last_name, u.email
FROM project_members pm
JOIN users u ON pm.user_id = u.id
WHERE pm.project_id = $1
ORDER BY pm.created_at ASC;

-- name: UpdateProjectMemberRole :one
UPDATE project_members
SET role = $3
WHERE project_id = $1 AND user_id = $2
RETURNING *;

-- name: RemoveProjectMember :exec
DELETE FROM project_members
WHERE project_id = $1 AND user_id = $2;
