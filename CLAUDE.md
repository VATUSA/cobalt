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

## Event banner uploads (DigitalOcean Spaces)

Event banners used to be caller-supplied URLs, which in practice meant Imgur (blocked in
several jurisdictions, including the UK) or Discord (which discourages off-app CDN use).
We now host them ourselves in the **`vatusa-events` Spaces bucket, region `sfo3`** — note
that most other VATUSA buckets are `nyc3`.

`POST /event/create` and `POST /event/:id` accept **either** a JSON body (unchanged
contract: `banner_image_url` is used as-is) **or** `multipart/form-data` with the same
fields plus a `banner_image` file part. The staff app in `webapps` always sends multipart.
When a file is present, `storage.UploadEventBanner` validates and uploads it and the
generated URL is what lands in `events.banner_image_url` — the database schema is
unchanged. When no file is present, `banner_image_url` from the request is kept, which is
what lets an edit that doesn't replace the image keep its existing banner.

The upload happens **after** the facility permission check, so an unauthorised request
can never write to the bucket. Uploads are validated in `storage/image.go`: max 8 MB, must
decode as PNG/JPEG/GIF/WebP, must be 16:9 within a 0.02 tolerance. The decode is
security-relevant — without it the endpoint would store HTML or SVG on a public bucket and
hand back a URL, which is a stored-XSS primitive. Object keys embed 16 random bytes and
are never rewritten, so objects are immutable and cached with `max-age=31536000`; replacing
a banner writes a new object and orphans the old one.

`storage/sigv4.go` signs the single PutObject call by hand rather than pulling in the AWS
SDK (~15 modules, plus its default flexible-checksum behaviour has to be disabled for
non-AWS S3 endpoints). It's pinned against AWS's published test vectors in
`storage/sigv4_test.go`.

Config lives in `config/spaces.go` (`DO_SPACES_KEY`, `DO_SPACES_SECRET`, `DO_SPACES_REGION`,
`DO_SPACES_BUCKET`, optional `DO_SPACES_ENDPOINT` and `DO_SPACES_PUBLIC_BASE_URL`). With no
key/secret set — the normal local setup — uploads return `storage.ErrNotConfigured` and only
the JSON/URL path works.

Policy documents (see below) reuse this same pattern against a **second bucket**,
`vatusa-storage` in region `nyc3`, sharing `DO_SPACES_KEY`/`DO_SPACES_SECRET` but with its
own `DO_SPACES_DOCS_REGION`/`DO_SPACES_DOCS_BUCKET`/`DO_SPACES_DOCS_ENDPOINT`/
`DO_SPACES_DOCS_PUBLIC_BASE_URL` in `config/spaces.go`. `storage/document.go` validates and
uploads via the same hand-rolled SigV4 signer in `storage/sigv4.go`; concurrent uploads
across both buckets share a package-level 4-slot semaphore in `storage/spaces.go` since each
upload holds the whole file in memory twice (multipart read, then SigV4 payload hash).

## FAQ, solo certs, and policy documents

Three simple CRUD table groups, each following the `event`/`news` pattern of a public read
endpoint plus `AssertGlobal`/`AssertFacility`-gated writes:

- **FAQ** (`/faq`) — `faq_category` and `faq_item`, global-write only (`acl.ObjectFaq`).
- **Solo certs** (`/solo`) — `solo_cert`, one active cert per `(cid, position)` enforced at
  the application level (MySQL can't express a filtered unique index). Write access is
  scoped per-controller via `endpoints.AssertFacilityForCid`, which resolves the caller's
  permission against the target controller's home *and* visiting facilities and returns
  whichever facility actually granted access — that's the facility stamped onto the record,
  and edits/deletes are scoped to the record's own `facility` column (not the controller's
  possibly-since-transferred current facility).
- **Policy** (`/policy`) — `policy_category` and `policy_document`, global-write only
  (`acl.ObjectPolicy`). `policy_document.document_url` is validated with
  `config.IsSafeDocumentURL` (https-only, same scheme rule as `IsAllowedRedirect`) since it's
  rendered as an `href` in the staff app.

All three follow `event`'s multipart-or-JSON upload pattern (`isMultipartRequest` in
`endpoints/params.go`, shared with `event.go`), stamp `created_by_cid`/`updated_by_cid` from
`auth.GetUserCid(c)`, and return dates/timestamps as formatted strings
(`time.DateOnly`/`config.TimestampFormat`), not raw `time.Time`, matching `models/Event.go`.

## API testing

A **Bruno** collection lives in `Bruno/` (api, login, roster, user, Events, News, etc.) for
exercising endpoints locally.
