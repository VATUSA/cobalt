-- LEGACY-CONTROLLERS-PATCH: temporary support for `current` (legacy Laravel site), which
-- still authenticates via `vatusa-old.controllers`. Delete this file and its Go caller
-- (grep LEGACY-CONTROLLERS-PATCH) once `current` no longer depends on that table.

-- name: ProvisionLegacyController :exec
INSERT IGNORE INTO `vatusa-old`.controllers
    (cid, fname, lname, email, facility, rating, created_at, updated_at,
     facility_join, flag_homecontroller, lastactivity, prefname)
VALUES
    (?, ?, ?, ?, ?, ?, NOW(), NOW(), NOW(), ?, NOW(), 0);
