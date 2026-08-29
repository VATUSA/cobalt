# config.IsAllowedRedirect (config/env.go) gates the caller-supplied `redirect`
# query param against REDIRECT_ALLOWLIST=vatusa.test (docker-compose.test.yml)
# to prevent open redirects. Disallowed targets never even reach VATSIM as
# `state` (validatedRedirect returns "" in endpoints/login.go); allowed ones
# are carried through and honored on /login/connect.

POST http://localhost:9100/mock/set-scenario
{
    "scenario": "active"
}
HTTP 204

# Disallowed redirect target is dropped before it ever leaves cobalt -- final
# destination falls back to POST_LOGIN_URL, same as no redirect at all.
GET http://localhost:8080/login?redirect=https://evil.example.com
[Options]
location: true
HTTP 200
[Asserts]
body == "900001"

# Allowed redirect target (matches REDIRECT_ALLOWLIST) is carried through
# VATSIM's `state` round trip and honored. vatusa.test doesn't resolve, so
# this follows the chain by hand (capturing each Location) instead of an
# automatic `location: true` follow, which would error trying to actually
# connect to that last hop.
GET http://localhost:8080/login?redirect=https://vatusa.test/ok
HTTP 302
[Captures]
# cobalt's ConnectAuthorizeParams (vatsim/connect.go) puts a literal, unescaped
# space in the `scope` param -- invalid to reuse directly as a URL, so escape it.
authorize_url: header "Location" replace " " "%20"

GET {{authorize_url}}
HTTP 302
[Captures]
connect_url: header "Location"

GET {{connect_url}}
HTTP 302
[Asserts]
header "Location" == "https://vatusa.test/ok"
