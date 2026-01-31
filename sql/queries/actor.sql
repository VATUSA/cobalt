-- name: GetActiveActorTokens :many
SELECT
    actor.id,
    actor.name,
    actor.actor_type,
    actor.rate_limit_override,
    actor.rate_limit_bypass,
    actor_token.comment,
    actor_token.token
FROM actor_token
    JOIN actor ON actor_token.actor_id = actor.id
WHERE actor.is_active AND actor_token.is_active;

-- name: GetACLGrantsForActor :many
SELECT *
FROM actor_acl
WHERE actor_id = ?;