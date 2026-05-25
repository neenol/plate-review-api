# PlateReview API

A Go HTTP API for the PlateReview app. PlateReview is a PWA where users look up US license plates and read/write reviews about drivers. This repo is the backend; frontend is in `plate-review-web`.

## Stack

- Go 1.26
- chi router
- PostgreSQL (Neon for dev/prod)
- sqlc for type-safe queries
- goose for migrations
- Clerk for auth (we verify JWTs issued by Clerk)
- Deployed to Fly.io

## Architecture

The codebase is organized into three layers — **controller**, **service**, **repository / platform** — with strict, one-directional dependencies. **Controllers depend on services. Services depend on repositories and platform clients. Lower layers never call upward.**

### Controller layer (`internal/controllers/`)

- All HTTP-aware code lives here and nowhere else.
- Responsibilities: parse and validate request input, call a service method, translate the result (or error) into an HTTP response.
- Controllers know about `http.Request`, `http.ResponseWriter`, JSON, status codes, headers.
- Controllers do NOT contain business logic, do NOT touch the database, do NOT call third parties directly.
- One controller file per resource (e.g. `reviews.go`, `users.go`, `plates.go`).
- Controllers receive their service dependencies via struct fields, set up at wiring time.

### Service layer (`internal/services/`)

- All business logic lives here.
- Responsibilities: orchestrate the work needed to fulfill a use case — validation that goes beyond input shape (e.g. uniqueness checks), normalization, authorization decisions, calling repositories and platform clients in the right order, mapping domain errors.
- Services take and return **domain types** defined in `internal/domain/`. They do NOT take or return HTTP types, request DTOs, or sqlc-generated types.
- Services depend on **interfaces** for their dependencies (repository interfaces, platform client interfaces). This keeps services testable with fakes and decouples them from the concrete implementations.
- One service per logical area (e.g. `ReviewService`, `UserService`, `PlateService`, `VoteService`).

### Repository layer (`internal/repositories/`)

- All database access lives here.
- Responsibilities: execute SQL queries (via sqlc), map sqlc row types to domain types, handle DB-specific errors (e.g. unique violation → domain error).
- Repositories expose interfaces that services depend on. The concrete implementation wraps sqlc's generated `Queries` type.
- Repositories do NOT contain business logic.
- One repository per aggregate (e.g. `UserRepository`, `PlateRepository`, `ReviewRepository`, `ReplyRepository`, `VoteRepository`).

### Platform layer (`internal/platform/`)

- All third-party / external service code lives here.
- For this project: a Clerk client that verifies JWTs and (if needed later) calls Clerk's REST API.
- Same pattern as repositories: an interface consumed by services, with a concrete implementation here.
- One package per third party (e.g. `internal/platform/clerk/`).

### Domain types (`internal/domain/`)

- Plain Go structs representing core entities: `User`, `Plate`, `Review`, `Reply`, `Vote`.
- Domain-level errors (e.g. `ErrNotFound`, `ErrUsernameTaken`, `ErrUnauthorized`) live here too.
- No tags for JSON or database — those concerns belong to controllers and repositories respectively.

### Dependency wiring (`internal/server/`)

- `server.go` constructs concrete implementations (repository structs, platform clients) and wires them into services, then wires services into controllers, then mounts controllers on the chi router.
- This is the only place that knows about every layer at once. Everything else only knows its direct dependencies.

## Layering rules (strict)

- Controllers → Services → Repositories / Platform → External world
- Services depend on **interfaces** defined alongside the service (e.g. `services/review_service.go` declares the `ReviewRepository` interface it needs).
- Repositories and platform clients **implement** those interfaces; they don't know about services.
- No circular dependencies, ever.
- Domain types are usable from any layer; nothing else crosses layers.

If you find yourself wanting to import `net/http` in a service, stop — that logic belongs in a controller. If you find yourself wanting to write SQL in a service, stop — that belongs in a repository.

## Project conventions

- **No ORMs.** Use sqlc. Write SQL in `internal/db/queries/*.sql`, run `sqlc generate`, use the generated code from the repository layer only.
- **No global state.** Dependencies are passed explicitly through structs.
- **Errors:** return errors up the stack with `fmt.Errorf("context: %w", err)`. Services translate infrastructure errors into domain errors (`ErrNotFound`, `ErrConflict`, etc.). Controllers translate domain errors into HTTP responses via a centralized helper.
- **Logging:** structured with `slog`. Include `request_id` in every log line via middleware.
- **Tests:**
  - Service tests use fake implementations of the repository/platform interfaces — fast, no DB required.
  - Repository tests hit a real Postgres (testcontainers-go or a dedicated test DB).
  - Controller tests use `httptest` against the chi router with a fake service.
  - Plate normalization and hashing are pure functions — straightforward table-driven tests.
- **No premature interfaces.** Define interfaces only where there's a real second implementation (real + fake for tests counts). Don't make every struct an interface.
- **Imports:** stdlib, blank line, third-party, blank line, internal. `goimports` enforces this.

## Project structure

```
plate-review-api/
  cmd/server/
    main.go              -- entry point: load config, build server, listen
  internal/
    config/              -- env loading
    domain/              -- domain types and domain errors
      user.go
      plate.go
      review.go
      reply.go
      vote.go
      errors.go
    controllers/         -- HTTP handlers; one file per resource
      users.go
      plates.go
      reviews.go
      replies.go
      votes.go
      errors.go          -- domain error → HTTP response mapping
      request.go         -- request DTOs and parsing helpers
    services/            -- business logic
      user_service.go
      plate_service.go
      review_service.go
      reply_service.go
      vote_service.go
    repositories/        -- DB access; implementations of service-defined interfaces
      user_repository.go
      plate_repository.go
      review_repository.go
      reply_repository.go
      vote_repository.go
    platform/
      clerk/             -- Clerk JWT verification + API client
    db/
      queries/           -- *.sql files for sqlc
      migrations/        -- goose migrations
      sqlc.yaml
      generated/         -- sqlc output (committed)
    middleware/          -- auth context injection, logging, CORS, request ID
    plate/               -- normalization + hashing (pure functions; used by service layer)
    server/
      server.go          -- wires everything together, returns http.Handler
  go.mod
  go.sum
  .env.example
  Dockerfile
  fly.toml
  Makefile
  CLAUDE.md
  README.md
```

## Request flow example

A `POST /plates/CA/ABC123/reviews` request:

1. **chi router** → routes to the controllers' `reviews.Create` handler.
2. **Middleware** has already verified the Clerk JWT and put the Clerk user ID into the request context.
3. **Controller** parses the JSON body into a request DTO, validates input shape, extracts the user ID and plate params, calls `reviewService.Create(ctx, params)`.
4. **Service** looks up the user via `userRepo.GetByClerkID`, normalizes and hashes the plate via the `plate` package, ensures the plate row exists via `plateRepo.GetOrCreate`, inserts the review via `reviewRepo.Create`, returns a `domain.Review`.
5. **Controller** serializes the domain review into a JSON response with the right status code.
6. If any step returned an error, the controller maps it: `ErrNotFound` → 404, `ErrUsernameTaken` → 409, `ErrUnauthorized` → 403, anything else → 500 (logged).

## Data model

See `internal/db/migrations/` for source of truth. Summary:

- `users`: our user row, linked to a `clerk_user_id`. Has a unique `username`.
- `plates`: keyed by `plate_hash = sha256(state + ":" + normalized + ":" + pepper)`. We never store raw plate numbers.
- `reviews`: text reviews attached to a plate.
- `replies`: nested replies on reviews (self-referential via `parent_reply_id`).
- `votes`: polymorphic upvotes on reviews or replies. POC is upvote-only (value=+1).

## Plate normalization

Before hashing a plate:
1. Uppercase
2. Strip non-alphanumeric
3. Validate state against the 50-state + DC list
4. Reject empty or >8 chars

Hash: `sha256(state + ":" + normalized + ":" + PLATE_PEPPER)`. `PLATE_PEPPER` is an env var. This logic lives in `internal/plate/` and is called from the service layer, never directly from controllers.

## API contract

See README. Standard error shape: `{ "error": { "code": "...", "message": "..." } }`. Auth via `Authorization: Bearer <clerk-jwt>` header.

## Environment variables

See `.env.example`. Required:
- `DATABASE_URL`
- `CLERK_SECRET_KEY`
- `CLERK_JWT_KEY` (for offline JWT verification)
- `PLATE_PEPPER`
- `PORT` (default 8080)
- `CORS_ORIGIN` (frontend URL)

## Common commands

- `make dev` — run with air for live reload
- `make test` — run unit tests
- `make migrate-up` — apply migrations
- `make migrate-new name=add_foo` — create a new migration
- `make sqlc` — regenerate DB code

## Out of scope for POC

- Moderation tooling (report button stub OK, no admin panel)
- Notifications
- Editing reviews/replies (delete + repost is fine)
- Downvotes
- Search/discovery features
- Internationalization (US-only for now)