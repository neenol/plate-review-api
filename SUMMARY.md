# Build Summary

## What was built

A complete Go HTTP API scaffold for PlateReview with every layer implemented and tested.

### Structure delivered

| Layer | Package | Files |
|---|---|---|
| Entry point | `cmd/server` | `main.go` |
| Config | `internal/config` | `config.go` |
| Domain | `internal/domain` | `user`, `plate`, `review`, `reply`, `vote`, `errors` |
| Plate utils | `internal/plate` | `plate.go` + `plate_test.go` |
| DB migrations | `internal/db/migrations` | `001_initial.sql` |
| sqlc queries | `internal/db/queries` | 5 `.sql` files |
| Generated DB code | `internal/db/generated` | sqlc output (committed) |
| Repositories | `internal/repositories` | 5 concrete implementations |
| Platform | `internal/platform/clerk` | offline RSA JWT verifier |
| Services | `internal/services` | 5 services + `deps.go` for interfaces |
| Controllers | `internal/controllers` | 5 controllers + error/request helpers |
| Middleware | `internal/middleware` | RequestID, Logger, RequireAuth, OptionalAuth |
| Server wiring | `internal/server` | `server.go` constructs the full dependency graph |

### Endpoints

```
GET  /health
POST /users                           (auth required)
GET  /users/me                        (auth required)
PATCH /users/me                       (auth required)
GET  /plates/{state}/{plate}
GET  /plates/{state}/{plate}/reviews
POST /plates/{state}/{plate}/reviews  (auth required)
DELETE /reviews/{id}                  (auth required, owner only)
GET  /reviews/{id}/replies
POST /reviews/{id}/replies            (auth required)
GET  /replies/{id}/replies
POST /replies/{id}/replies            (auth required)
DELETE /replies/{id}                  (auth required, owner only)
POST /reviews/{id}/votes              (auth required)
DELETE /reviews/{id}/votes            (auth required)
POST /replies/{id}/votes              (auth required)
DELETE /replies/{id}/votes            (auth required)
```

### Tests

- **`internal/plate`** — 10 table-driven tests for Normalize / Hash / NormalizeAndHash
- **`internal/services`** — UserService, ReviewService, VoteService tested against in-memory fakes; no DB required
- **`internal/controllers`** — UserController, ReviewController tested with httptest + stub services
- **`internal/repositories`** — Stubbed with TODO; no DB wired yet (see Next Steps)

All pass: `go test ./...` ✓  
Build: `go build ./...` ✓  
Vet: `go vet ./...` ✓

## Key decisions

See `DECISIONS.md` for full rationale.  Short version:

- **`internal/services/deps.go`** — all repository/platform interfaces live in one file to avoid duplicate declarations when multiple services share an interface (e.g. `UserRepository`)
- **Manual Clerk JWT verification** — uses `golang-jwt/jwt/v5` + RSA public key; avoids the heavy official SDK
- **`score` via JOIN** — vote counts are computed in the listing SQL query to avoid N+1
- **Plate `GetOrCreate` race** — handled with insert-then-retry-read on conflict
- **Polymorphic votes** — `target_type + target_id` on a single `votes` table

## Stubbed / deferred

| Item | Location | Note |
|---|---|---|
| Repository integration tests | `internal/repositories/repositories_test.go` | TODO comment; needs testcontainers-go or a real test DB |
| `CLERK_SECRET_KEY` usage | `internal/platform/clerk` | Imported in config and passed to server; not yet used (needed when calling Clerk's REST API, e.g. user deletion) |
| Reply service tests | `internal/services/` | Only UserService, ReviewService, VoteService have tests; ReplyService follows the same pattern |
| Report / moderation stub | — | Out of scope per CLAUDE.md POC definition |

## Next steps

### Required before the server can start

1. **Provision Neon database**
   - Create a project at neon.tech
   - Copy the connection string

2. **Create a Clerk application**
   - Dashboard → API Keys → copy `Secret key` and `JWT public key`

3. **Fill in `.env`**
   ```sh
   cp .env.example .env
   # Edit DATABASE_URL, CLERK_SECRET_KEY, CLERK_JWT_KEY, PLATE_PEPPER, CORS_ORIGIN
   ```
   > `PLATE_PEPPER` must be a strong random value (e.g. `openssl rand -hex 32`).
   > **Never change it after plates are in the database** — all existing hashes will break.

4. **Run migrations**
   ```sh
   source .env
   make migrate-up
   ```

5. **Start the server**
   ```sh
   make dev    # with air (live reload)
   # or
   make build && ./bin/server
   ```

### Optional next steps

6. **Wire up repository integration tests** — add `testcontainers-go` to the test setup in `internal/repositories/`.

7. **Deploy to Fly.io**
   ```sh
   fly launch   # (if not already set up)
   fly secrets set DATABASE_URL=... CLERK_SECRET_KEY=... CLERK_JWT_KEY=... PLATE_PEPPER=...
   fly deploy
   ```

8. **Add `CORS_ORIGIN`** to the Fly secrets pointing at the production frontend URL.

9. **Install `air`** for live reload during development:
   ```sh
   go install github.com/air-verse/air@latest
   ```
