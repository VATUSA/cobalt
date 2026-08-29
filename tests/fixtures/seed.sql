-- Runs once cobalt's own migrations have created the schema (see
-- docker-compose.test.yml's `seed` service). Inserts rows nothing in the
-- login path creates itself.

-- `user` rows for the mock IdP's canned CIDs (see cmd/mockidp/main.go).
-- Connect() upserts `vatsim_user` itself but never creates `user` -- that's
-- normally done by the async vatsim_sync job, which is skipped outside
-- prod/staging (endpoints/login.go). GetCombinedUserByCID inner-joins the two,
-- so without this row the login callback 500s with "user record does not
-- exist" for every scenario.
INSERT INTO user (cid, display_name, facility) VALUES
    (900001, 'Mock Active', 'ZHQ'),
    (900002, 'Mock Suspended', 'ZHQ');

-- One actor with the system_internal role (mapped to
-- ObjectLegacyLoginToken read+write in acl/role_config.go), used by
-- token_session.hurl to exercise the actor-gated GET /token/:cid and
-- POST /tokenSession endpoints.
INSERT INTO actor (id, name, actor_type, is_active, created_at, created_by_cid)
VALUES (1, 'integration-test-actor', 'system_internal', 1, UNIX_TIMESTAMP(), 900001);

INSERT INTO acl_actor_role (actor_id, facility, role, grantor_cid, granted_at)
VALUES (1, '*', 'system_internal', 900001, UNIX_TIMESTAMP());

INSERT INTO actor_token (actor_id, comment, token, is_active, created_at, created_by_cid, updated_at, updated_by_cid)
VALUES (1, 'integration test token', 'integration-test-actor-token', 1, UNIX_TIMESTAMP(), 900001, UNIX_TIMESTAMP(), 900001);
