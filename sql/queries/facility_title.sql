-- name: GetFacilityTitles :many
SELECT * FROM facility_title
WHERE facility = ?
ORDER BY id ASC;

-- name: GetAllFacilityTitles :many
SELECT id, facility, code
FROM facility_title;

-- name: GetFacilityTitleById :one
SELECT * FROM facility_title
WHERE id = ?;

-- name: CreateFacilityTitle :execresult
INSERT INTO facility_title (facility, title, code, created_at)
VALUES (?, ?, ?, ?);

-- name: DeleteFacilityTitleById :exec
DELETE FROM facility_title
WHERE id = ?;

-- name: GetUserTitlesByFacility :many
SELECT ft.id, ft.facility, ft.title, ft.code, ut.grantor_cid, ut.granted_at
FROM facility_title ft
JOIN user_title ut ON ut.title_id = ft.id
WHERE ut.cid = ? AND ft.facility = ?
ORDER BY ft.id ASC;

-- name: AssignUserTitle :exec
INSERT IGNORE INTO user_title (cid, title_id, grantor_cid, granted_at)
VALUES (?, ?, ?, ?);

-- name: DeleteUserTitle :exec
DELETE FROM user_title
WHERE cid = ? AND title_id = ?;
