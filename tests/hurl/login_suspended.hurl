# Connect() rejects a suspended/inactive VATSIM rating (endpoints/login.go)
# before ever minting a session cookie.

POST http://localhost:9100/mock/set-scenario
{
    "scenario": "suspended"
}
HTTP 204

GET http://localhost:8080/login
[Options]
location: true
HTTP 403
[Asserts]
cookie "vatusa-cobalt-token" not exists
