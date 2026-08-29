# GET /token/:cid and POST /tokenSession are actor-gated (middleware/auth_actor.go)
# rather than cookie-gated -- used by the legacy site and the prod->staging
# relay to mint/redeem a session token server-to-server. Exercises the actor
# token seeded in tests/fixtures/seed.sql.
#
# middleware.NewActorAuth loads the active-token cache once at cobalt startup
# and only refreshes it once a minute (auth/token.go), and the seed step
# necessarily runs after cobalt is already healthy -- so the seeded token can
# take up to ~60s to become visible. The first request below retries for that.

GET http://localhost:8080/token/900001
X-Auth-Token: integration-test-actor-token
[Options]
retry: 40
retry-interval: 2000
HTTP 200
[Captures]
minted_token: jsonpath "$.token"

POST http://localhost:8080/tokenSession
X-Auth-Token: integration-test-actor-token
Content-Type: application/json
{
    "token": "{{minted_token}}"
}
HTTP 200
[Asserts]
jsonpath "$.user.cid" == 900001

# Without a valid actor token, both endpoints are forbidden.
GET http://localhost:8080/token/900001
HTTP 403

POST http://localhost:8080/tokenSession
Content-Type: application/json
{
    "token": "{{minted_token}}"
}
HTTP 403
