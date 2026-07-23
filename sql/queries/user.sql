-- name: GetUserByCID :one
SELECT u.*,
       (SELECT GROUP_CONCAT(facility SEPARATOR ',') from user_visit uv where uv.cid = u.cid) as visiting_facilities
FROM user u
WHERE u.cid = ?;


-- name: GetCombinedUser :many
SELECT vu.cid
     , vu.name_first
     , vu.name_last
     , vu.email
     , vu.rating
     , vu.pilotrating
     , vu.militaryrating
     , vu.region_id
     , vu.division_id
     , vu.subdivision_id
     , vu.latest_rating_change
     , vu.last_sync
     , u.display_name
     , u.controller_rating
     , u.instructor_rating
     , u.facility
     , (SELECT GROUP_CONCAT(facility SEPARATOR ',') from user_visit uv where uv.cid = u.cid) as visiting_facilities
     , u.discord_id
     , u.last_promotion_time
     , u.last_transfer_time
     , u.last_visit_time
     , u.last_competency_date
from vatsim_user vu
         join user u on vu.cid = u.cid
where (sqlc.narg(cid) is null OR vu.cid = sqlc.narg(cid))
AND (sqlc.narg(has_cids_slice) IS NULL OR vu.cid IN (sqlc.slice(cids)))
AND (sqlc.narg(home_facility) is null or u.facility = sqlc.narg(home_facility))
AND (sqlc.narg(visit_facility) is null or u.cid IN (select cid from user_visit uv where uv.facility = sqlc.narg(visit_facility)));

-- name: UpsertUserForMigration :exec
INSERT INTO user (cid, display_name, controller_rating, instructor_rating, facility, discord_id,
                  last_promotion_time, last_transfer_time, last_competency_date)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE display_name         = IF(user.display_name = '', VALUES(display_name), user.display_name),
                        controller_rating    = VALUES(controller_rating),
                        instructor_rating    = VALUES(instructor_rating),
                        facility             = VALUES(facility),
                        discord_id           = VALUES(discord_id),
                        last_promotion_time  = VALUES(last_promotion_time),
                        last_transfer_time   = VALUES(last_transfer_time),
                        last_competency_date = VALUES(last_competency_date);

-- name: InsertUserFromVatsimSync :exec
INSERT IGNORE INTO user (cid, display_name, controller_rating, instructor_rating, facility)
VALUES (?, ?, 0, 0, ?);

-- name: UpdateUserFromVatsimSync :exec
UPDATE user SET display_name = ? WHERE cid = ?;


-- name: UpdateUserForTransfer :exec
UPDATE user SET facility = ?, last_transfer_time = ? WHERE cid = ?;

-- name: GetUserRatingHours :one
SELECT *
FROM user_rating_hours
WHERE cid = ? AND rating = ?;

-- name: StoreUserRatingHours :exec
INSERT INTO user_rating_hours (cid, rating, hours, last_check_time) VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE hours = VALUES(hours), last_check_time = VALUES(last_check_time);

-- name: GetUserRatingHourCheckCids :many
SELECT u.cid
FROM user u
JOIN vatsim_user vu on u.cid = vu.cid
LEFT JOIN user_rating_hours h ON u.cid = h.cid AND h.rating = least(u.controller_rating, 5)
WHERE (h.hours IS NULL or h.hours < ?)
AND u.controller_rating > 1
AND (u.controller_rating >= 4 OR (vu.region_id = 'AMAS' AND vu.division_id = 'USA'));