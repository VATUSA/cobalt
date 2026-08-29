# GET /login/logout (endpoints/login.go GetLogout) clears the session cookie
# and redirects to POST_LOGIN_URL.

POST http://localhost:9100/mock/set-scenario
{
    "scenario": "active"
}
HTTP 204

GET http://localhost:8080/login
[Options]
location: true
HTTP 200
[Asserts]
body == "900001"

GET http://localhost:8080/login/logout
HTTP 302
[Asserts]
header "Location" == "http://localhost:8080/login/whoami"
header "Set-Cookie" contains "vatusa-cobalt-token="
header "Set-Cookie" contains "Max-Age=0"

GET http://localhost:8080/login/whoami
HTTP 200
[Asserts]
body == "-1"
