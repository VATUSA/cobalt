# Roles, Permissions, and Titles

Cobalt actually has **three separate, loosely-connected concepts** that all get
casually called "roles" in conversation. This doc pulls them apart:

1. **Role** — a string tag on a user (or API-key actor) that expands, via a
   static table, into a set of **Permissions**. This is what authorization
   checks actually run against.
2. **Permission** — an `(Object, Action)` pair, e.g. `write:event`. Not stored
   per-grant anywhere; derived at request time from a subject's Roles.
3. **Title** — a facility-defined, purely cosmetic label ("Air Traffic
   Manager", "Mentor") shown on rosters. Holding a Title grants **zero**
   permissions by itself.

The confusion mostly comes from the fact that the legacy VATUSA system had one
flat idea ("this person is the ATM of ZNY") and cobalt splits that into two
independent records (a Role grant *and* a Title grant) fed by two independent
migration jobs, plus some historical dead code sitting next to the live code.
Both are called out below.

---

## 1. Role

A `Role` (`src/acl/role_config.go`) is just a `type Role string` — a bare
constant, not a struct. Roles are granted to a *subject* at a *facility
scope* and stored in one of two tables:

```sql
-- migration 000006 — human users, keyed by VATSIM CID
CREATE TABLE acl_user_role (
    id bigint PRIMARY KEY,
    cid int not null,
    facility varchar(4) not null,   -- real facility code, or '*' for global
    role varchar(120) not null,
    grantor_cid int not null,
    granted_at bigint not null
);

-- migration 000007 — API-key "actors" (service/machine auth)
CREATE TABLE acl_actor_role (
    id bigint PRIMARY KEY,
    actor_id int not null,
    facility varchar(4) not null,
    role varchar(120) not null,
    grantor_cid int not null,
    granted_at bigint not null
);
```

`acl_user_role` is what `GrantRole`/`RevokeRole` (`src/endpoints/roles.go`)
manage through the HTTP API. **`acl_actor_role` has no HTTP surface at all** —
grep finds no caller of `AddRoleToActor`/`RemoveRoleFromActor` outside the
generated `db` package, so API-key role grants are provisioned out-of-band
(direct DB writes), not through the app.

### Scoping

Both tables key roles by `(subject, facility, role)`. `facility` is either a
real code (`"ZNY"`) or the sentinel `acl.ScopedRoleGlobalFacility` (`"*"`),
meaning "granted division-wide." The DB doesn't enforce which roles are
legal at which scope — the app does, in `validateRoleFacility` (see §4).

Every human additionally gets two roles synthesized in memory, never stored:

- `RoleAuthenticatedUser` — anyone with a session, scope `*`.
- `RoleDivisionMember` — anyone with `DivisionID == "USA"`, scope `*`.

These two are collectively `acl.AutomaticRoles`, and they're the *only* two
roles excluded from "is this person staff" (see the `IsStaffRole` gotcha in
§7).

### Full role list (`role_config.go`)

`system_administrator`, `authenticated_user`, `division_member`,
`division_staff`, `division_management`, `division_tech_team`,
`division_command_center`, `social_media_team`, `ace_team`, `email_user`,
`air_traffic_manager`, `deputy_air_traffic_manager`,
`training_administrator`, `event_coordinator`, `facility_engineer`,
`web_maintainer`, `instructor`, `mentor`, `division_academy_editor`,
`facility_academy_editor`, `system_internal`, `system_facility`,
`system_external`.

---

## 2. Permission

Permissions are **not stored anywhere** — there is no `permissions` table.
They're a compile-time mapping from Role → a list of `(Object, Action)`
pairs, defined in `src/acl/role_config.go`:

```go
var RoleGlobalPermissions   = map[Role][]PermissionDefinition{ ... }
var RoleFacilityPermissions = map[Role][]PermissionDefinition{ ... }
```

`Action` is one of `read`, `write`, `manage_unowned`, `usage`. `Object` is a
business resource: `superadmin`, `news_post`, `event`, `event_approval`,
`faq`, `solo_cert`, `policy`,
`user_sensitive_details`, `roster`, `facility_tech_config`,
`facility_title`, `facility_title_management`,
`facility_title_senior_staff`, `facility_title_junior_staff`,
`system_api_role`, `division_management_role`, `division_staff_role`,
`facility_senior_staff_role`, `facility_junior_staff_role`,
`facility_training_role`, `legacy_role_sync`, `legacy_login_token`.

`ActionUsage` on `ObjectSuperAdmin` is a hardcoded escape hatch: holding it
makes every other global permission check pass automatically
(`acl/handler.go`), which is how `RoleSystemAdmin` works.

The last six `*_role` objects plus `legacy_role_sync`/`legacy_login_token`
are `acl.MaskedObjects` — stripped out of anything sent to the client (JWT
claims, `/my/session` response). This matters for the webapps side (§8).

### Resolving roles into permissions: `PermissionHandler`

`acl.PermissionHandler` (`src/acl/handler.go` — the name is misleading, it's
**not** an HTTP handler, it's the in-memory permission set built from a
subject's roles) is constructed once per request from `[]ScopedRole` and
answers `HasGlobal(object, action)` / `HasFacility(facility, object, action)`.

The one genuinely surprising rule: **if a normally facility-scoped role is
granted with `Facility: "*"`, its facility permissions get promoted to
global** — i.e. they apply everywhere, not nowhere. Confirmed by
`TestPermissionHandler_GlobalRoleAssignmentGrantsEveryFacility` in
`handler_test.go`. This is intentional (a way to make e.g. a global ATM-like
grant) but is not obvious from the types.

`HasFacility` also checks `HasGlobal` first — a global grant on an object
satisfies any facility, including ones that don't exist, since facility
existence isn't re-checked at this layer.

### Assigning roles: a separate governing map

Who is allowed to *grant/revoke* a given role is a different question from
what that role's holder can do, governed by:

```go
var RoleToPermissionObjectMap = map[Role]Object{
    RoleSystemAdmin:        ObjectSystemRole,
    RoleDivisionStaff:      ObjectDivisionManagementRole,
    RoleDivisionManagement: ObjectDivisionManagementRole,
    RoleAirTrafficManager:  ObjectFacilitySeniorStaffRole,
    RoleEventCoordinator:   ObjectFacilityJuniorStaffRole,
    RoleMentor:             ObjectFacilityTrainingRole,
    // ...
}
```

Rule: to grant/revoke role `X`, you need `write` on
`RoleToPermissionObjectMap[X]` at the relevant scope. This is what makes
`GetAssignableRoles` (§4) and `GrantRole`/`RevokeRole` (§5) work.

---

## 3. Title (`FacilityTitle` / `UserFacilityTitle`)

A completely separate system, added in migration `000021`:

```sql
CREATE TABLE facility_title (
    id bigint PRIMARY KEY,
    facility varchar(4) not null,
    title varchar(128) not null,          -- "Air Traffic Manager"
    code varchar(4) not null default '',  -- "ATM"
    tier varchar(6) not null default '',  -- 'senior' | 'junior'
    UNIQUE (facility, title)
);

CREATE TABLE user_title (
    id bigint PRIMARY KEY,
    cid bigint not null,
    title_id bigint not null,
    grantor_cid int not null,
    granted_at bigint not null,
    UNIQUE (cid, title_id)
);
```

Every facility is seeded with one `facility_title` row per legacy title code
(`ATM`/`DATM`/`TA`/`EC`/`FE`/`WM` — tier `senior` for the first four, `junior`
for `EC`/`FE`/`WM` — plus `US1`–`US9` at `ZHQ`).

**A Title, by itself, grants no permissions.** `FacilityTitle` and
`UserFacilityTitle` (`src/models/FacilityTitle.go`) are plain DTOs with no
authorization logic attached. Titles exist purely as a roster display label.

The only place `tier` matters for authorization is deciding **who can
assign/revoke a title of that tier** — not what the title-holder can do:

```go
var TitleTierToPermissionObjectMap = map[TitleTier]Object{
    TitleTierSenior: ObjectFacilityTitleSeniorStaff,
    TitleTierJunior: ObjectFacilityTitleJuniorStaff,
}
```

`AssignUserFacilityTitles` (`src/endpoints/facility_title.go`) resolves the
title's tier and requires `write` on the corresponding object at that
facility. Notably, ATM/DATM/TA hold `write:facility_title_junior_staff`
at their own facility but **not** the senior-staff object — so a facility's
own senior staff cannot grant/revoke senior titles at their own facility;
only Division Management (which holds both tier objects globally) can.

### Titles vs. Roles are populated independently

Both `SyncRolesForUser` (writes `acl_user_role`) and `BulkMigrateTitles`
(writes `user_title`) consume the *same* legacy role codes (`ATM`, `DATM`,
etc.), but they are **two unrelated, non-transactional jobs** with no shared
state or FK between the resulting rows. Granting the `RoleAirTrafficManager`
role does not create the "Air Traffic Manager" title and vice versa — they
can drift independently. This dual-write from one legacy source is very
likely the biggest single reason the system feels confusing: what used to be
one fact in the legacy system ("this person is ATM of ZNY") is now two
separate facts in cobalt that happen to usually agree.

---

## 4. LegacyRole — three artifacts, only one of them live

There are three separate "legacy role mapping" things in the codebase.

**A. `src/models/LegacyRole.go`** — the wire DTO for the sync API:

```go
type LegacyRole struct {
    Role     string `json:"role"`
    Facility string `json:"facility"`
}
type SyncRolesRequest struct {
    CID   int          `json:"cid"`
    Roles []LegacyRole `json:"roles"`
}
```

This is what the legacy PHP system POSTs to `POST /roles/legacy_sync` (single
user) or `POST /roles/legacy_sync/bulk`, gated by
`write:legacy_role_sync` globally — only `RoleSystemInternal` (a service
actor) has that permission.

**B. `src/legacy_migration/roles.go`: `LegacyToModernRoleMap`** — the mapping
that actually runs:

```go
var LegacyToModernRoleMap = map[string]acl.Role{
    "ATM": acl.RoleAirTrafficManager, "DATM": acl.RoleDeputyAirTrafficManager,
    "TA": acl.RoleTrainingAdministrator, "EC": acl.RoleEventCoordinator,
    "FE": acl.RoleFacilityEngineer, "WM": acl.RoleWebMaintainer,
    "INS": acl.RoleInstructor, "MTR": acl.RoleMentor,
    "CBT": acl.RoleDivisionAcademyEditor, "FACCBT": acl.RoleFacilityAcademyEditor,
    "ACE": acl.RoleACETeam, "DCC": acl.RoleDivisionCommandCenter,
    "SMT": acl.RoleSocialMediaTeam, "USWT": acl.RoleDivisionTechTeam,
    "EMAIL": acl.RoleEmailUser,
}
```

Plus a fallback for `US*`-prefixed codes not in the map: any `US*` code
grants `RoleDivisionStaff`; `US1`–`US4` additionally grant
`RoleDivisionManagement`; `US0`/`US6` additionally grant `RoleSystemAdmin`.
Unrecognized, non-`US`-prefixed codes are silently dropped.

`SyncRolesForUser` converts an incoming legacy role list through this map,
diffs it against the user's current `acl_user_role` rows
(`DetermineDropAddRoles`), and issues adds/removes to converge. **This is a
full reconciliation, not an additive sync** — the legacy system is the
source of truth, and `acl_user_role` is a derived mirror of it.
`BulkMigrateRoles` does the same thing for every user, for backfills.

**C. `src/acl/role_legacy_config.go`: `LegacyRoleToRolesMap` /
`LegacyGlobalRoles`** — **dead code.** A structurally similar but not
identical third mapping (e.g. it hardcodes `"US0"` → a flat role list rather
than B's layered `US*`-prefix logic, and has a `"DICE": {}` entry found
nowhere else). A repo-wide grep confirms zero references to either symbol
outside their own declaration file. It sits in the `acl` package right next
to the live `role_config.go`, so it looks load-bearing and isn't. Treat any
reasoning based on this file as wrong; the map in (B) is the truth. This is
a good candidate for deletion.

---

## 5. How a request actually gets authorized

There is **no ACL middleware** in `src/routes/routes.go`. Every route is
registered plainly; the only blanket middleware resolves *identity*
(`actorAuth.Middleware` for API keys, `CookieAuth` for user sessions), not
*authorization*. Permission checks are imperative, inline, at the top of
each handler, via helpers in `src/endpoints/acl_context_funcs.go`:

```go
func AssertGlobal(c *echo.Context, object acl.Object, action acl.Action) bool {
    if HasGlobal(c, object, action) { return true }
    _ = RespondError(c, http.StatusForbidden, ...)
    return false
}
func AssertFacility(c *echo.Context, facility string, object acl.Object, action acl.Action) bool { ... }
```

A handler calls `if !AssertGlobal(c, ...) { return nil }` and bails if it
fails (the assert already wrote the 403). There's no single place to read
off what a route requires — you have to open the handler. Three examples:

**Global, service-only** — `LegacySyncRoles`:
```go
if !AssertGlobal(c, acl.ObjectLegacyRoleSync, acl.ActionWrite) { return nil }
```

**Facility-scoped, dynamic object** — `AssignUserFacilityTitles` resolves
the required object from the title's tier at request time
(`resolveTitleTierPermission`) before calling `AssertFacility`.

**Mixed global-or-facility** — `GrantRole`/`RevokeRole` route through
`checkRoleManagePerm`:
```go
func checkRoleManagePerm(c *echo.Context, facility string, object acl.Object) bool {
    if facility == "ZHQ" { return AssertGlobal(c, object, acl.ActionWrite) }
    return AssertFacility(c, facility, object, acl.ActionWrite)
}
```

**Scoped to a specific controller, not a fixed facility** — solo certs don't have
a facility known up front the way an event's `facility` field does; the
facility to check is the *target controller's own*, and it may be their home
facility or one they're only visiting. `AssertFacilityForCid` handles this:

```go
func AssertFacilityForCid(c *echo.Context, targetCid int, object acl.Object, action acl.Action) (string, bool)
```

It looks up `targetCid` via `dbconn.GetCombinedUserByCID`, then checks (in
order) a global grant, a facility grant on the controller's home facility, and
a facility grant on each of the controller's visiting facilities — returning
whichever facility actually granted access, which is what the caller stamps
onto the record being written (e.g. `solo_cert.facility`). A caller with no
grant of `object:action` at *any* scope (checked via `acl.HasAny`, hoisted
above the CID lookup) gets a flat 403 without a DB round trip and without a
404-vs-403 split that would otherwise let an unprivileged caller enumerate
which CIDs exist.

Note `"ZHQ"` here — the HTTP/URL layer uses the literal facility code
`"ZHQ"` to mean "division-wide," while the DB/ACL layer uses the sentinel
`"*"` (`acl.ScopedRoleGlobalFacility`). `GrantRole` explicitly translates
between the two (`dbFacility := facility; if facility == "ZHQ" { dbFacility
= acl.ScopedRoleGlobalFacility }`). Nothing enforces this translation
uniformly elsewhere — e.g. `GetAssignableRoles` (§6) returns `"ZHQ"`
directly as a map key. **Two different spellings of "global" exist across
layers**; keep this in mind when tracing a bug across the HTTP boundary.

---

## 6. Who can assign which roles: `GetAssignableRoles`

```go
func (ph *PermissionHandler) GetAssignableRoles() map[string][]Role {
    // for every role in RoleToPermissionObjectMap:
    //   if I have global write on its governing object -> bucket under "ZHQ"
    //   else, for each facility where I hold *any* facility permission,
    //     if I have facility-write on its governing object there -> bucket under that facility
}
```

This is **not** an explicit rank/hierarchy check — it's purely "do I hold
`write` on the role's governing `Object`, at what scope." It only looks
hierarchical because `RoleToPermissionObjectMap` /
`RoleGlobalPermissions` / `RoleFacilityPermissions` were hand-authored to
form one: Division Management can write both `facility_senior_staff_role`
and `facility_junior_staff_role`, but an ATM can only write
`facility_junior_staff_role` + `facility_training_role` — so an ATM can
grant Mentor/EC/FE/WM but not another ATM/DATM/TA.

Exposed over HTTP as `GET /my/roles/assignable` →
`models.AssignableRoles{Roles: map[string][]string}` — this is exactly the
shape webapps' `CobaltAssignableRoles` consumes (§8).

---

## 7. Caching (`src/acl/cache.go`)

`PermissionHandlerCache` is a **process-local, in-memory** map (keyed by CID
for users, actor id for API keys) with a flat **1-minute TTL**
(`IsStale()`). There is no invalidation on write: `GrantRole`/`RevokeRole`/
legacy sync do not bust the cache. After a role change, that user's own
in-process cached permission set can be up to 60 seconds stale — and since
this cache is per-pod with no shared/distributed layer, a multi-pod
deployment doesn't even share a staleness window; each pod expires
independently.

### The staff-ness gotcha

`IsStaffRole(role)` is defined as "not one of the two `AutomaticRoles`" —
meaning a garbage/typo'd role string in the DB grants **zero permissions**
but still flips `IsStaff()` to `true` (confirmed by
`TestPermissionHandler_UnknownRoleGrantsNothing`). Fails open on
"staff-ness," fails closed on actual permissions.

---

## 8. Cross-check against webapps

`packages/third-party/src/cobalt.ts` is the API client and carries **two
parallel representations of the same permission data**, because cobalt
emits permissions differently depending on the source:

- `/my/session` JSON response → array-of-objects
  (`CobaltPermission { action, object, facility? }`), consumed by
  `normalizePermissionCollections` in `apps/staff/lib/acl.ts`.
- JWT cookie claims → comma-separated `"object:action"` /
  `"facility:object:action"` strings (mirroring
  `PermissionHandler.GetGlobalPermissionsString()` /
  `GetFacilityPermissionsString()` in `acl/handler.go`), parsed by
  `transformCobaltJwt` in `cobalt.ts`.

These two paths are not unified into one shape before use — worth knowing if
you're debugging "the UI shows different permissions than the API."

`apps/staff/lib/acl.ts` re-declares its own `ACTION`/`OBJECT` constants
mirroring `acl.Action`/`acl.Object`, by hand, with no shared schema or
codegen between Go and TS. It's currently **missing** `facility_title`,
`facility_title_management`, `facility_title_senior_staff`,
`facility_title_junior_staff`, `roster`, `facility_tech_config`,
`legacy_role_sync`, `legacy_login_token`. Harmless today since nothing
references them, but a trap for whoever builds title-management UI next.

`apps/staff/lib/assignableRoles.ts` correctly consumes the server's
`"ZHQ"`-keyed division-wide bucket from §6 (`GLOBAL_ROLE_FACILITY = "ZHQ"`
matches the endpoint-layer convention), and the only role-gated nav logic
actually wired up in the sidebar (`AppSideBar.tsx`) checks `ObjectEvent`/
`ObjectNewsPost`/the ACE-team assignable-role check — none of which are
masked, so that path is sound.

**Dead client-side code worth flagging**: `apps/staff/lib/acl.ts` also
defines `hasAnyStaffAccess()` and `buildStaffSidebarCapabilities()`, which
check `OBJECT.systemApiRole`, `OBJECT.divisionManagementRole`,
`OBJECT.divisionStaffRole`, `OBJECT.facilitySeniorStaffRole`,
`OBJECT.facilityJuniorStaffRole`, `OBJECT.facilityTrainingRole` — but
**every one of these six objects is in cobalt's `MaskedObjects` list**
(`acl/permissions.go`), so they are stripped from both the JSON response and
the JWT before the client ever sees them. If either function is ever wired
up, it will silently and permanently evaluate false for these checks — e.g.
a pure Mentor/Instructor/Facility-Academy-Editor holder has *no*
non-masked permissions at all, so this logic would report them as having no
staff access despite legitimately holding a staff role. Confirmed neither
function is currently called anywhere in webapps, so this isn't live, but
it's exactly the kind of code someone will reach for next and be quietly
misled by, since the masking rationale isn't visible from the TS side.

---

## 9. Summary: known sharp edges

| # | Issue | Where |
|---|-------|-------|
| 1 | `role_legacy_config.go`'s `LegacyRoleToRolesMap`/`LegacyGlobalRoles` are unused, disagree with the live mapping | `src/acl/role_legacy_config.go` (dead; live logic is `src/legacy_migration/roles.go`) |
| 2 | Two spellings of "global": `"ZHQ"` at the HTTP layer, `"*"` at the DB/ACL layer | `endpoints/roles.go`, `acl.ScopedRoleGlobalFacility` |
| 3 | A facility-scoped role granted with `Facility: "*"` becomes global everywhere, not nowhere | `acl/handler.go`, `NewPermissionHandler` |
| 4 | Unknown role strings grant no permissions but still count as "staff" | `acl/role_config.go`, `IsStaffRole` |
| 5 | No endpoint grants/revokes `acl_actor_role` — API-key roles are DB-manual | `src/db/acl_actor_role.sql.go` |
| 6 | Titles and Roles are two independent, non-transactional systems fed from the same legacy source — they can drift | `legacy_migration/roles.go` vs `background/migrate_titles.go` |
| 7 | Permission cache has a flat 1-minute TTL and no invalidation on write | `src/acl/cache.go` |
| 8 | `PermissionHandler`/`handler.go` names suggest an HTTP handler; it's actually the in-memory permission-set evaluator | `src/acl/handler.go` |
| 9 | webapps' `acl.ts` OBJECT/ACTION constants are hand-mirrored from Go and currently incomplete | `webapps/apps/staff/lib/acl.ts` |
| 10 | webapps has dead sidebar-capability code that checks masked (never-sent) permission objects | `webapps/apps/staff/lib/acl.ts`, `hasAnyStaffAccess`/`buildStaffSidebarCapabilities` |
