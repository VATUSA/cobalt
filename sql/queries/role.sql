-- name: FetchGlobalRolesByCID :many
SELECT *
FROM role_global
WHERE cid = ?;

-- name: FetchFacilityRolesByCID :many
SELECT *
FROM role_facility
WHERE cid = ?;

-- name: CreateGlobalRole :exec
INSERT INTO role_global (cid, role, created_at, created_by)
VALUES(?, ?, ?, ?);

-- name: CreateFacilityRole :exec
INSERT INTO role_facility (cid, role, facility, created_at, created_by)
VALUES(?, ?, ?, ?, ?);

-- name: DeleteUserGlobalRoles :exec
DELETE FROM role_global WHERE cid = ?;

-- name: DeleteUserFacilityRoles :exec
DELETE FROM role_facility WHERE cid = ?;