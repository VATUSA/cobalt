# cobalt-api

Newer VATUSA backend API, intended to extend/replace the legacy Laravel API (`api`). It's the
sole identity provider for the VATUSA stack — `webapps` (the modern frontend) and `current`
(the main website) both delegate login to it.

## Stack

- **Go 1.25**, HTTP via [Echo v5](https://github.com/labstack/echo)
- **MySQL**, accessed through [sqlc](https://sqlc.dev)-generated code (`src/db`, generated
  from `sql/queries` + `sql/migrations` + `sql/legacy_schema`) — never hand-edit `src/db`
- Migrations applied with [golang-migrate](https://github.com/golang-migrate/migrate)
- JWT sessions via `golang-jwt`
- VATSIM Connect (OAuth2) for login, plus VATSIM API v2 for roster/rating data
- `k8s.io/client-go` for cluster access from within the app (e.g. background jobs)

Three binaries are built from `src/cmd/`: `server` (HTTP API), `background` (async worker),
`cli` (one-off tasks). The Dockerfile builds all three; its entrypoint defaults to `server`.

## Setup

```sh
cp .env.example .env   # fill in DB credentials, JWT_KEY, etc.
cd src
go build ./...
go run ./cmd/server.go
```

Run tests with `go test ./...` from `src/`.

If you change a query in `sql/queries` or the schema, regenerate the DB layer from the repo
root:

```sh
sqlc generate
```

## Project layout

| Package | Purpose |
|---|---|
| `endpoints`, `routes`, `middleware` | HTTP layer |
| `auth`, `login`, `acl` | Authentication (JWT) and role/permission-based access control |
| `db`, `dbconn`, `models` | Data access (`db` is sqlc-generated) |
| `action`, `roster`, `background` | Domain logic and async work |
| `vatsim` | VATSIM API/Connect integration |
| `moodle` | Moodle integration (e.g. rating cohort sync) |
| `legacy_migration` | Migration helpers from the legacy (VATUSA-old) system |
| `config`, `cli`, `server`, `cmd` | Wiring and entrypoints |

## Routes

All routes are registered in `src/routes/routes.go`.

| Group | Endpoints | Purpose |
|---|---|---|
| `/token`, `/tokenSession` | `GET /token/:cid`, `POST /tokenSession` | Actor-authenticated token issuance/lookup (server-to-server) |
| `/news` | latest/page/post CRUD | News posts |
| `/roles` | `POST /legacy_sync[/bulk]` | Sync roles from the legacy system |
| `/event` | upcoming/page/CRUD/review | Events |
| `/user/:cid` | `GET`, `GET /blockers` | User lookup |
| `/login` | see below | VATSIM Connect login flow |
| `/my` | `GET /session`, `POST /transfer` | Current-user session info, facility transfer requests |
| `/roster/:facility` | roster + pending transfers | Facility roster management |
| `/facility/:facility` | `GET /v3/apikeys` | Facility API key/tech config |

## Authentication & authorization

Two independent auth mechanisms are layered on every request (`routes.SetupRoutes`):

1. **Actor auth** (`middleware.ActorAuth` / `auth.ActorTokenHeader`) — a static bearer token
   presented by trusted server-to-server callers (e.g. `current`, or a relaying instance of
   cobalt itself). Resolves to an actor ID cached from the DB and refreshed every minute.
2. **Cookie auth** (`middleware.CookieAuth`) — reads the JWT session cookie
   (`auth.JWTCookieName`) set at login and resolves it to a user CID.

Handlers that require a logged-in user wrap with `middleware.RequireLogin`. Authorization on
top of that is role/permission-based (`acl` package): roles are scoped either globally or to a
facility, and `acl.PermissionHandler` checks an actor's or user's cached roles against
`(object, action)` or `(facility, object, action)` permission keys.

### Login / VATSIM Connect flow

Cobalt owns the VATSIM Connect OAuth registration, and VATSIM only whitelists a
`redirect_uri` on **cobalt.vatusa.net** — so every non-prod (dev/staging) login has to proxy
through the prod instance:

- `GET /login` — on non-prod, redirects to prod's `/login/staging` (VATSIM would reject a
  non-prod redirect_uri).
- `GET /login/staging` (prod-only) — logs into prod via VATSIM Connect if needed, then calls
  back to the originating dev/staging instance's `/token/:cid` (actor-token gated) to mint a
  token, and redirects the browser to that instance's `/login/useToken/:token`.
- `GET /login/connect` — VATSIM's OAuth callback; sets the session cookie. If prod is mid-relay
  for a dev/staging login, loops back into `/login/staging` to complete the handoff instead of
  landing on `POST_LOGIN_URL`.
- `GET /login/useToken/:token` (non-prod only) — sets the session cookie from the relayed
  token.

A caller-supplied `redirect` query param (validated against `REDIRECT_ALLOWLIST` via
`config.IsAllowedRedirect`) lets consumers like `current` request their own callback instead
of the default `POST_LOGIN_URL`; it's carried across the VATSIM round trip via the OAuth
`state` param. A `_staging_relay|`-prefixed `state` is how `Connect()` distinguishes a relayed
dev/staging login from an ordinary prod login.

See `CLAUDE.md` for more implementation detail on this flow.

## API testing

A [Bruno](https://www.usebruno.com) collection lives in `Bruno/` for exercising endpoints
locally (api, login, roster, user, events, news, etc.).
