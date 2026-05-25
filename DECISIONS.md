# Architecture & Design Decisions

## Interface placement (`internal/services/deps.go`)

**Decision:** All repository and platform interfaces consumed by services live in
`internal/services/deps.go` rather than scattered across individual service files.

**Rationale:** The spec says "interfaces belong with the consumer," which is the
services package.  Multiple services share `UserRepository` (review, reply, vote
services all need it).  Declaring the same interface name in two files of the same
package is a compile error in Go.  A single `deps.go` satisfies the spirit of the
rule (interfaces live in the package that consumes them) without duplicate
declarations.

---

## Clerk JWT verification — manual RSA, not SDK

**Decision:** Clerk JWTs are verified offline using `golang-jwt/jwt/v5` with the
RSA public key from `CLERK_JWT_KEY`.  The official `clerk-sdk-go` is not imported.

**Rationale:** The only Clerk capability needed for the POC is JWT verification.
The official SDK pulls in a large dependency tree.  The spec explicitly provides
`CLERK_JWT_KEY` for offline verification.  A thin wrapper around `golang-jwt` is
simpler, faster in tests, and has no network dependency.

---

## `PlateService` is separate from `ReviewService`

**Decision:** `PlateService.GetOrCreate` is a standalone service.  `ReviewService`
replicates a private `getOrCreatePlate` helper for the create-review path.

**Rationale:** The spec lists `PlateService` as a distinct service.  Controllers
call `PlateController.Get` for plate metadata.  The `ReviewService` must also
get-or-create a plate internally; rather than wiring PlateService into
ReviewService (coupling two services), the plate lookup logic uses the raw
repository interfaces that ReviewService already holds.  The duplication is small
and avoids service→service calls which the spec forbids.

---

## Vote polymorphism via `target_type` + `target_id`

**Decision:** The `votes` table stores `target_type IN ('review','reply')` and
`target_id UUID` without foreign-key constraints, plus a unique index on
`(user_id, target_type, target_id)`.

**Rationale:** The spec explicitly describes polymorphic upvotes.  Separate FK
columns would require two nullable columns with a CHECK constraint — equivalent
complexity.  The current schema is the idiomatic POC approach; a CHECK on
`target_type` keeps values safe.  Referential integrity can be added later via a
trigger or by splitting to two tables.

---

## Plate `GetOrCreate` race handling

**Decision:** Both `PlateService.GetOrCreate` and `ReviewService.getOrCreatePlate`
use a read-then-write with a conflict-retry fallback: on `ErrConflict` from
`Create`, fall back to `GetByHash`.

**Rationale:** The application has no distributed lock.  Under concurrent requests
for the same new plate, two goroutines may both find `ErrNotFound` and race to
insert.  The insert that loses the race gets a unique violation (`ErrConflict`);
the retry read succeeds.  This is safe because `plate_hash` is immutable once
created.

---

## `Score` included in list queries via JOIN

**Decision:** `ListReviewsByPlate` and the reply list queries compute `score` via
`COALESCE(SUM(v.value), 0)::bigint` in SQL rather than a separate round-trip per
row.

**Rationale:** Avoids N+1 queries.  The POC has no pagination so full-table
aggregation per plate is acceptable.  If performance becomes a concern, a
materialized column updated by a trigger is the next step.

---

## Timestamp format in JSON responses

**Decision:** All `created_at` / `updated_at` fields are serialized as
`"2006-01-02T15:04:05Z"` (RFC 3339 UTC, second precision).

**Rationale:** Consistent with how most JS clients parse dates.  Sub-second
precision is not needed for the POC.

---

## `CORS_ORIGIN` defaults to `"*"` if unset

**Decision:** When `CORS_ORIGIN` is empty the server allows all origins.

**Rationale:** `CORS_ORIGIN` is not in the required-variable check because the
server can function without it during local development.  Production deployments
must set it explicitly to the frontend URL.
