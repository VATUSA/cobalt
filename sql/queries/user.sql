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
     , u.last_competency_date
from vatsim_user vu
         join user u on vu.cid = u.cid
where (sqlc.narg(cid) is null OR vu.cid = sqlc.narg(cid))
AND (sqlc.narg(home_facility) is null or u.facility = sqlc.narg(home_facility))
AND (sqlc.narg(visit_facility) is null or u.cid IN (select cid from user_visit uv where uv.facility = sqlc.narg(visit_facility)));

-- name: UpsertUserForMigration :exec
INSERT INTO user (cid, controller_rating, instructor_rating, facility, discord_id, last_promotion_time,
                  last_transfer_time, last_competency_date)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE controller_rating    = VALUES(controller_rating),
                        instructor_rating    = VALUES(instructor_rating),
                        facility             = VALUES(facility),
                        discord_id           = VALUES(discord_id),
                        last_promotion_time  = VALUES(last_promotion_time),
                        last_transfer_time   = VALUES(last_transfer_time),
                        last_competency_date = VALUES(last_competency_date);

-- name: InsertUserFromVatsimSync :exec
INSERT IGNORE INTO user (cid, display_name, controller_rating, instructor_rating, facility)
VALUES (?, ?, 0, 0, ?);


-- name: UpdateUserForTransfer :exec
UPDATE user SET facility = ?, last_transfer_time = ? WHERE cid = ?;