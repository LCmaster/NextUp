-- Backfill project_members for any project whose owner has no entry yet.
-- This repairs projects created before the atomic transaction fix was applied.
INSERT INTO project_members (project_id, user_id, role)
SELECT p.id, p.owner_id, 'owner'
FROM   projects p
WHERE  p.deleted_at IS NULL
  AND  NOT EXISTS (
       SELECT 1 FROM project_members pm
       WHERE  pm.project_id = p.id
         AND  pm.user_id    = p.owner_id
  )
ON CONFLICT (project_id, user_id) DO NOTHING;
