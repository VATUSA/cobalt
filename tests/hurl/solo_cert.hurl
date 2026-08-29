# /solo CRUD: the AssertFacilityForCid enumeration-oracle fix (an anonymous
# caller must get 403 for both a real and a fake cid, never a 404 that would
# leak which), that the stamped facility comes from the target controller's
# own record rather than the caller's, the duplicate-active-cert 409, and 404
# on operations against a missing cert.
#
# The only role grantable to a loginable identity in this harness
# (RoleDivisionStaff, granted globally to 900001 in seed.sql) carries a
# *global* write:solo_cert, so every create below resolves through
# AssertFacilityForCid's HasGlobal branch. The other branch it has -- a
# facility-scoped caller matching one of the target's *visiting* facilities,
# returning that facility rather than the target's home one -- needs a
# facility-scoped (not global) grant to exercise, which would take a second
# independently-scoped loginable identity this harness doesn't have; that
# path is covered by TestPermissionHandler_HasFacility-style reasoning over
# the code plus the manual spot-check in the plan this test suite was
# written against, not by this file.
#
# expires_soon/expires_later are passed in via `just test-integration`
# (computed relative to the real clock, since expires validation is checked
# against it).

# Before any session exists, a caller with no acl grant at any scope must be
# rejected outright -- for both a cid that exists (900001) and one that
# doesn't (999999999). Both must come back 403; if the 404 branch in
# AssertFacilityForCid ran before the permission check, only the fake cid
# would 404, letting an anonymous caller enumerate real cids one probe at a
# time.
POST http://localhost:8080/solo
{
    "cid": 900001,
    "position": "ZHQ_APP",
    "expires": "{{expires_soon}}"
}
HTTP 403

POST http://localhost:8080/solo
{
    "cid": 999999999,
    "position": "ZHQ_APP",
    "expires": "{{expires_soon}}"
}
HTTP 403

POST http://localhost:9100/mock/set-scenario
{
    "scenario": "active"
}
HTTP 204

# 900001 is granted the global "division_staff" role in seed.sql, which
# carries write:solo_cert.
GET http://localhost:8080/login
[Options]
location: true
HTTP 200

# 900010 (seed.sql) is home ZAB, not 900001's own ZHQ -- created and checked
# alone (nothing else in solo_cert yet) so the facility on the response can
# be asserted unambiguously: it must be the target's own home facility, not
# the caller's.
POST http://localhost:8080/solo
{
    "cid": 900010,
    "position": "ZAB_APP",
    "expires": "{{expires_soon}}"
}
HTTP 200
[Captures]
other_cert_id: jsonpath "$.id"

GET http://localhost:8080/solo
HTTP 200
[Asserts]
jsonpath "$" count == 1
jsonpath "$[0].id" == {{other_cert_id}}
jsonpath "$[0].facility" == "ZAB"

POST http://localhost:8080/solo
{
    "cid": 900001,
    "position": "ZHQ_APP",
    "expires": "{{expires_soon}}"
}
HTTP 200
[Captures]
cert_id: jsonpath "$.id"

GET http://localhost:8080/solo
HTTP 200
[Asserts]
jsonpath "$" count == 2

# A second active cert for the same (cid, position) is a 409 -- checked at
# the application level since MySQL can't express a filtered unique index.
POST http://localhost:8080/solo
{
    "cid": 900001,
    "position": "ZHQ_APP",
    "expires": "{{expires_later}}"
}
HTTP 409

# A different position for the same cid is unaffected by the check above.
POST http://localhost:8080/solo
{
    "cid": 900001,
    "position": "ZHQ_TWR",
    "expires": "{{expires_soon}}"
}
HTTP 200
[Captures]
twr_cert_id: jsonpath "$.id"

# expires validation: a clearly-past date and a clearly-more-than-45-days
# date are both rejected regardless of when this test happens to run.
POST http://localhost:8080/solo
{
    "cid": 900001,
    "position": "ZHQ_GND",
    "expires": "2020-01-01"
}
HTTP 400

POST http://localhost:8080/solo
{
    "cid": 900001,
    "position": "ZHQ_GND",
    "expires": "2099-01-01"
}
HTTP 400

POST http://localhost:8080/solo
{
    "cid": 900001,
    "position": "ZHQ_GND",
    "expires": "not-a-date"
}
HTTP 400

# UpdateSoloCert no longer accepts/rewrites facility -- it's scoped and
# stamped once, at creation.
POST http://localhost:8080/solo/{{cert_id}}
{
    "position": "ZHQ_APP",
    "expires": "{{expires_later}}"
}
HTTP 200

# Updating/deleting a nonexistent id is a real 404.
POST http://localhost:8080/solo/999999
{
    "position": "ZHQ_APP",
    "expires": "{{expires_soon}}"
}
HTTP 404

DELETE http://localhost:8080/solo/999999
HTTP 404

DELETE http://localhost:8080/solo/{{cert_id}}
HTTP 200

DELETE http://localhost:8080/solo/{{other_cert_id}}
HTTP 200

DELETE http://localhost:8080/solo/{{twr_cert_id}}
HTTP 200
