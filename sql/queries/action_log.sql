-- name: CreateActionLog :exec
INSERT INTO action_log (actor_cid, target_cid, action, log, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetActionLogsForTarget :many
SELECT * FROM action_log WHERE target_cid = ? ORDER BY created_at DESC;
