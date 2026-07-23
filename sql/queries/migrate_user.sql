-- name: GetLegacyControllerByCID :one
SELECT c.cid, c.facility, c.rating, c.discord_id
FROM `vatusa-old`.controllers c
where c.cid = ?;

-- name: GetLegacyLastPromotionDateByCID :one
SELECT cid, cast(max(created_at) as datetime) as last_promotion_date
from `vatusa-old`.promotions
where cid = ?
  and `to` <= 5
  and `from` < 5
group by cid;

-- name: GetLegacyLastTransferDateByCID :one
SELECT cid, cast(max(created_at) as datetime) as last_transfer_date
from `vatusa-old`.transfers
WHERE cid = ?
  and status = 1
group by cid;

-- name: GetLegacyLastCompetencyDateByCID :one
SELECT cid, cast(max(competency_date) as datetime) as last_competency_date
from `vatusa-old`.controller_eligibility_cache
where competency_date is not null and cid = ?
group by cid;

-- name: GetLegacyControllersToMigrate :many
SELECT c.cid, c.facility, c.rating, c.discord_id, TRIM(CONCAT_WS(' ', c.fname, c.lname)) as display_name
from `vatusa-old`.controllers c
where c.rating > 0
  AND (
    c.rating > 1 OR c.facility NOT IN ('ZAE', 'ZZN', 'ZZI')
    );

-- name: GetLegacyLastPromotionDatesToMigrate :many
SELECT cid, cast(max(created_at) as datetime) as last_promotion_date
from `vatusa-old`.promotions
where `to` <= 5
  and `from` < 5
group by cid;

-- name: GetLegacyLastTransferDateToMigrate :many
SELECT cid, cast(max(created_at) as datetime) as last_transfer_date
from `vatusa-old`.transfers
WHERE status = 1
group by cid;

-- name: GetLegacyLastCompetencyDateToMigrate :many
SELECT cid, cast(max(competency_date) as datetime) as last_competency_date
from `vatusa-old`.controller_eligibility_cache
where competency_date is not null
group by cid;

-- name: GetLegacyControllerRatingFromPromotionsToMigrate :many
SELECT p.cid, cast(max(`to`) as signed) as max_rating
FROM `vatusa-old`.promotions p
         JOIN `vatusa-old`.controllers c ON p.cid = c.cid
where `to` <= 7
  and c.rating > 5
group by p.cid;

-- name: GetLegacyInstructorTA :many
SELECT r.cid, r.role
FROM `vatusa-old`.roles r
JOIN `vatusa-old`.controllers c on r.cid = c.cid
where r.role = 'TA' or r.role = 'INS'
group by r.cid, r.role;

-- name: GetLegacyRoles :many
SELECT r.cid, r.role, r.facility
FROM `vatusa-old`.roles r;