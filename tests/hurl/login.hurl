# Happy-path VATSIM Connect login: GET /login redirects to the mock IdP's
# authorize endpoint, which auto-"consents" and bounces back to
# /login/connect, which sets the session cookie and lands on POST_LOGIN_URL.

POST http://localhost:9100/mock/set-scenario
{
    "scenario": "active"
}
HTTP 204

GET http://localhost:8080/login
HTTP 302
[Asserts]
header "Location" contains "localhost:9100/oauth/authorize"
header "Location" contains "client_id=test-client-id"

GET http://localhost:8080/login
[Options]
location: true
HTTP 200
[Asserts]
body == "900001"
