-- name: GetRolesForActor :many
SELECT *
FROM acl_actor_role
WHERE actor_id = ?;

-- name: AddRoleToActor :exec
INSERT INTO acl_actor_role (actor_id, facility, role, grantor_cid, granted_at)
VALUES (?, ?, ?, ?, ?);

-- name: RemoveRoleFromActor :exec
DELETE
FROM acl_actor_role
WHERE actor_id = ?
  AND facility = ?
  AND role = ?;