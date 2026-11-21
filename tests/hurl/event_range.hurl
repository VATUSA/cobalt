# GET /event/range (endpoints/event.go GetEventsInRange): the public calendar's
# feed. Unlike /event/upcoming it must return approved events that have already
# finished. Fixtures are the fixed January 2020 "Range Test" rows in seed.sql.

# Window [2020-01-01, 2020-02-01): the inside event plus both events that
# straddle an edge, in start order. The pending and March events are excluded.
GET http://localhost:8080/event/range?from=1577836800&to=1580515200
HTTP 200
[Asserts]
jsonpath "$" count == 3
jsonpath "$[0].title" == "Range Test Straddle Start"
jsonpath "$[1].title" == "Range Test Inside"
jsonpath "$[2].title" == "Range Test Straddle End"

# An event ending exactly at `from` does not overlap the window.
GET http://localhost:8080/event/range?from=1578625200&to=1578700000
HTTP 200
[Asserts]
jsonpath "$" count == 0

# Missing, malformed, empty/inverted and over-366-day windows are rejected.
GET http://localhost:8080/event/range?from=1577836800
HTTP 400

GET http://localhost:8080/event/range?from=abc&to=1580515200
HTTP 400

GET http://localhost:8080/event/range?from=1580515200&to=1577836800
HTTP 400

GET http://localhost:8080/event/range?from=1577836800&to=1609459201
HTTP 400

# /event/page/:page reads the page number from the path. A non-numeric or
# zero page is a 400 (it used to read ?page= and ignore the path entirely).
GET http://localhost:8080/event/page/abc
HTTP 400

GET http://localhost:8080/event/page/0
HTTP 400
