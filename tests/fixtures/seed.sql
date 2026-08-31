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

-- A global "division_staff" grant for the one loginable active mock CID
-- (900001), used by faq.hurl/solo_cert.hurl/policy.hurl to exercise the
-- authenticated write paths (acl.RoleDivisionStaff carries global write on
-- ObjectFaq/ObjectSoloCert/ObjectPolicy -- see acl/role_config.go).
INSERT INTO acl_user_role (cid, facility, role, grantor_cid, granted_at)
VALUES (900001, '*', 'division_staff', 900001, UNIX_TIMESTAMP());

-- A second controller, home facility ZAB (distinct from 900001's ZHQ), with
-- no login of its own -- used only as a solo_cert.hurl target to exercise
-- AssertFacilityForCid resolving a *different* facility than the caller's.
-- GetCombinedUserByCID inner-joins `user` with `vatsim_user` (only populated
-- for real logins via Connect()), so both rows are needed even though this
-- cid never logs in.
INSERT INTO vatsim_user (cid, name_first, name_last, email, rating, pilotrating, militaryrating, region_id, division_id, last_sync)
VALUES (900010, 'Mock', 'Other Facility', 'mock-other-facility@example.test', 5, 0, 0, 'AMAS', 'USA', UTC_TIMESTAMP());

INSERT INTO user (cid, display_name, facility) VALUES
    (900010, 'Mock Other Facility', 'ZAB');
