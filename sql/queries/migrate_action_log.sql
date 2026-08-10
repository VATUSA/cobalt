-- name: GetLegacyActionLogs :many
SELECT id, `from`, `to`, log, created_at, updated_at
FROM `vatusa-old`.action_log;
