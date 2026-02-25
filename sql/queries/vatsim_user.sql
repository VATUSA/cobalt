-- name: GetVatsimUserByCid :one
SELECT *
FROM vatsim_user
WHERE cid = ?;

-- name: UpsertVatsimUser :exec
INSERT INTO vatsim_user (cid, name_first, name_last, email, rating, pilotrating, militaryrating,
                         suspend_date, registration_date, region_id, division_id, subdivision_id, latest_rating_change,
                         last_sync)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE name_first           = VALUES(name_first),
                        name_last            = VALUES(name_last),
                        email                = VALUES(email),
                        rating               = VALUES(rating),
                        pilotrating          = VALUES(pilotrating),
                        militaryrating       = VALUES(militaryrating),
                        suspend_date         = VALUES(suspend_date),
                        registration_date    = VALUES(registration_date),
                        region_id            = VALUES(region_id),
                        division_id          = VALUES(division_id),
                        subdivision_id       = VALUES(subdivision_id),
                        latest_rating_change = VALUES(latest_rating_change),
                        last_sync            = VALUES(last_sync);