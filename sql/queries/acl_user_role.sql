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

-- name: GetUsersByRole :many
SELECT DISTINCT u.cid
     , u.display_name
     , u.controller_rating
     , u.instructor_rating
FROM acl_user_role aur
JOIN user u ON aur.cid = u.cid
WHERE aur.role = ?
  AND (sqlc.narg(facility) IS NULL OR aur.facility = sqlc.narg(facility));

-- name: GetFacilityStaffRoles :many
SELECT aur.role, u.cid, u.display_name
FROM acl_user_role aur
JOIN user u ON u.cid = aur.cid
WHERE aur.facility = ?
  AND aur.role IN (sqlc.slice(roles))
ORDER BY aur.role, u.display_name;
