-- name: CreateProjectInvite :one
INSERT INTO project_invites (project_id, email, token, role, expires_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (project_id, email) DO UPDATE SET token = EXCLUDED.token, role = EXCLUDED.role, expires_at = EXCLUDED.expires_at
RETURNING *;

-- name: GetProjectInviteByToken :one
SELECT * FROM project_invites
WHERE token = $1;

-- name: DeleteProjectInvite :exec
DELETE FROM project_invites
WHERE id = $1;

-- name: ListProjectInvites :many
SELECT * FROM project_invites
WHERE project_id = $1
ORDER BY created_at DESC;
