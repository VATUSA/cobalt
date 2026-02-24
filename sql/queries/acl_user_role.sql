-- name: GetRolesForUser :many
SELECT *
FROM acl_user_role
WHERE cid = ?;

-- name: AddRoleToUser :exec
INSERT INTO acl_user_role (cid, facility, role, grantor_cid, granted_at)
VALUES (?, ?, ?, ?, ?);

-- name: RemoveRoleFromUser :exec
DELETE
FROM acl_user_role
WHERE cid = ?
  AND facility = ?
  AND role = ?;