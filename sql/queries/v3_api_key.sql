-- name: GetV3ApiKeysByFacility :many
SELECT * FROM v3_api_key
WHERE facility = ? AND deleted_at IS NULL
ORDER BY testing ASC, id ASC;
