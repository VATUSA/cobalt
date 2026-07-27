# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this
repository. For how this repo relates to the other VATUSA projects, see the workspace
`CLAUDE.md` one directory up.

## Project Overview

**cobalt-api** — a newer VATUSA backend API written in Go (1.25). It is consumed by the
`webapps` frontend and by `current` (via that repo's `app/Cobalt` layer). Database
access is generated from SQL with **sqlc** (MySQL engine).

## Build & Development Commands

All Go code lives under `src/` (module `vatusa-cobalt`), so run Go commands from `src/`:

```sh
cd src
go build ./...
go test ./...
go run ./cmd/server.go        # HTTP API server
go run ./cmd/background.go     # Background worker
go run ./cmd/cli.go           # CLI tasks
```

### sqlc (generated DB layer)

`sqlc.yaml` (at repo root) generates the `db` package into `src/db` from:
- `sql/queries` — query definitions
- `sql/migrations` + `sql/legacy_schema` — schema

After editing SQL queries or schema, regenerate:

```sh
sqlc generate
```

Do **not** hand-edit files in `src/db` — they are generated. Migrations are applied via
`golang-migrate`.

## Architecture

Three binaries built from `src/cmd/` (`server`, `background`, `cli`). The Dockerfile builds
all three and defaults its entrypoint to `server`. Packages under `src/`:
- `endpoints`, `routes`, `middleware` — HTTP layer
- `auth`, `login`, `acl` — authentication and access control (JWT via `golang-jwt`)
- `db`, `dbconn`, `models` — data access (db is sqlc-generated)
- `action`, `roster`, `background` — domain logic and async work
- `vatsim` — VATSIM integration; `legacy_migration` — migration from the legacy system
- `config`, `cli`, `server`, `cmd` — wiring and entrypoints

## Login / VATSIM Connect flow

Cobalt is the sole identity provider for the VATUSA stack (`webapps` and `current` both
delegate login to it). It owns the VATSIM Connect OAuth registration — VATSIM only
whitelists a `redirect_uri` on **cobalt.vatusa.net**, so every hosted dev/staging login has
to proxy through the prod instance:

- `GET /login` (`endpoints/login.go: GetLogin`) — on a non-prod (`IsStaging()`) instance,
  redirects to prod's `/login/staging` instead of hitting VATSIM directly, since VATSIM would
  reject a non-prod redirect_uri.
- `GET /login/staging` (`GetLoginForStaging`, prod-only) — if not already logged into prod,
  kicks off VATSIM Connect on prod; once logged in, calls back to the origin dev/staging
  instance's internal `/token/:cid` (ACL-gated via `acl` + `middleware/auth_actor.go`'s
  actor-token cache) to mint a token, then redirects to that instance's
  `/login/useToken/:token`.
- `GET /login/connect` (`Connect`) — VATSIM's OAuth callback. Sets the session cookie. If
  this is prod runnning mid-relay (see below), loops back into `/login/staging` to complete
  the handoff instead of stopping at prod's own `POST_LOGIN_URL`.
- `GET /login/useToken/:token` (`LoginUseToken`, non-prod only) — sets the session cookie
  from the relayed token.

**Caller-supplied redirect target**: any of `GetLogin` / `GetLoginForStaging` /
`LoginUseToken` accept a `redirect` query param (e.g. so `current` can request its own
`/auth/callback` instead of the global default). It's validated by
`config.IsAllowedRedirect()` against the `REDIRECT_ALLOWLIST` env var (comma-separated host
suffixes) to prevent open redirects, then carried across the external VATSIM round trip via
the OAuth `state` param (echoed back verbatim by VATSIM on `/login/connect`) since there's no
other way to preserve it through that hop. When no `redirect` is supplied, behavior is
unchanged from before this existed — falls back to `config.PostLoginURL()`.

**`state` also marks relays**: because prod handles both its own logins and relayed
dev/staging ones, a redirect in `state` is not by itself evidence of a relay. Only
`GetLoginForStaging` tags its state, with the `_staging_relay|` prefix
(`stagingRelayState` / `parseStagingRelayState`), and `Connect()` loops back into
`/login/staging` for tagged state only. An untagged `state` is an ordinary login on this
instance and is honored directly. Treating any non-empty state on prod as a relay is what
sent prod legacy logins through the dev instance's `/token/:cid`, 500ing for every user
without a `cobalt_dev` row.

`current`'s side of this (the one-time `/auth/callback` handoff, decoupled from cobalt after
login) is documented in that repo's `CLAUDE.md`.

## API testing

A **Bruno** collection lives in `Bruno/` (api, login, roster, user, Events, News, etc.) for
exercising endpoints locally.
