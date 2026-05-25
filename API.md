# PlateReview API Contract

## Overview

Base URL: configured per environment (e.g. `https://plate-review-api.fly.dev` in production, `http://localhost:8080` in dev).

All request and response bodies are JSON. All timestamps are UTC ISO 8601 strings with second precision: `"2026-05-25T19:00:00Z"`. All IDs are UUIDs (lowercase hex with hyphens, e.g. `"a1b2c3d4-e5f6-..."`).

---

## Authentication

Protected endpoints require a valid Clerk-issued JWT in the `Authorization` header:

```
Authorization: Bearer <clerk_session_token>
```

Obtain the token from Clerk's frontend SDK (`session.getToken()`). Tokens are RS256-signed and verified offline by the server — no Clerk API call is made per request.

Endpoints marked **Auth required** return `401` if the header is missing or the token is expired/invalid.

---

## Common headers

| Header | Direction | Description |
|---|---|---|
| `Content-Type: application/json` | Request | Required on all POST/PATCH requests with a body |
| `Authorization: Bearer <token>` | Request | Required on protected endpoints |
| `X-Request-ID` | Response | Server-generated UUID for this request, present on every response |

---

## Error format

All errors use this shape regardless of status code:

```json
{
  "error": {
    "code": "not_found",
    "message": "resource not found"
  }
}
```

### Error codes

| `code` | HTTP status | Meaning |
|---|---|---|
| `unauthorized` | 401 | Missing or invalid JWT |
| `forbidden` | 403 | Authenticated but not allowed (e.g. deleting someone else's review) |
| `not_found` | 404 | Resource does not exist |
| `conflict` | 409 | Uniqueness violation (username taken, already voted, etc.) |
| `invalid_input` | 400 | Malformed JSON body, invalid UUID in path, failed plate validation |
| `internal_error` | 500 | Unexpected server error |

---

## Plate parameters

`{state}` and `{plate}` appear in several URLs. The server normalizes these before use:

- `{state}` — two-letter US state abbreviation or `DC`. Case-insensitive (`ca` and `CA` are both valid). Invalid state → `400 invalid_input`.
- `{plate}` — the raw plate number as it appears on the vehicle. The server strips non-alphanumeric characters and uppercases the remainder before hashing. Must produce a 1–8 character result after stripping, otherwise → `400 invalid_input`.

Plate data is stored by hash only; the raw number is never persisted.

---

## Endpoints

---

### Health

#### `GET /health`

Liveness check. No authentication required.

**Response `200 OK`**
```json
{ "status": "ok" }
```

---

### Users

User rows are created on the backend after a successful Clerk sign-up. The frontend must call `POST /users` once with the chosen username after the Clerk onboarding flow completes.

---

#### `POST /users`

Create a user record for the authenticated Clerk account.

**Auth required:** yes

**Request body**
```json
{
  "username": "alice"
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `username` | string | yes | Must be unique across all users |

**Response `201 Created`**
```json
{
  "id": "uuid",
  "username": "alice",
  "clerk_user_id": "user_2abc...",
  "created_at": "2026-05-25T19:00:00Z"
}
```

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | `username` is empty or body is malformed |
| 401 | `unauthorized` | Missing or invalid token |
| 409 | `conflict` | Username is already taken, or this Clerk account already has a user row |
| 500 | `internal_error` | Unexpected server error |

---

#### `GET /users/me`

Return the user record for the authenticated Clerk account.

**Auth required:** yes

**Response `200 OK`**
```json
{
  "id": "uuid",
  "username": "alice",
  "clerk_user_id": "user_2abc...",
  "created_at": "2026-05-25T19:00:00Z"
}
```

**Error responses**

| Status | Code | When |
|---|---|---|
| 401 | `unauthorized` | Missing or invalid token |
| 404 | `not_found` | No user row exists yet — frontend should redirect to onboarding |
| 500 | `internal_error` | Unexpected server error |

---

#### `PATCH /users/me`

Update the authenticated user's username.

**Auth required:** yes

**Request body**
```json
{
  "username": "alice_renamed"
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `username` | string | yes | Must be unique |

**Response `200 OK`** — same shape as `GET /users/me`.

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | `username` is empty or body is malformed |
| 401 | `unauthorized` | Missing or invalid token |
| 404 | `not_found` | Authenticated user has no user row |
| 409 | `conflict` | New username is already taken |
| 500 | `internal_error` | Unexpected server error |

---

### Plates

Plate records are created lazily — the first request for a given state/plate combination creates the row. Subsequent requests return the same record.

---

#### `GET /plates/{state}/{plate}`

Resolve (and if necessary create) the canonical plate record. Use this to check whether any reviews exist or to get the plate's internal ID before posting a review.

**Auth required:** no

**URL parameters**

| Param | Example | Notes |
|---|---|---|
| `state` | `CA` | 2-letter state code or `DC` (case-insensitive) |
| `plate` | `ABC123` | Raw plate string; special characters are stripped |

**Response `200 OK`**
```json
{
  "id": "uuid",
  "state": "CA",
  "created_at": "2026-05-25T19:00:00Z"
}
```

Note: `state` in the response is always uppercased. The raw plate number is not returned — it is never stored.

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | Invalid state, empty plate after stripping, or plate longer than 8 chars |
| 500 | `internal_error` | Unexpected server error |

---

### Reviews

---

#### `GET /plates/{state}/{plate}/reviews`

List all reviews for a plate, sorted newest-first.

**Auth required:** no

**URL parameters** — same as `GET /plates/{state}/{plate}`.

**Response `200 OK`** — array (empty array if no reviews exist; never 404):
```json
[
  {
    "id": "uuid",
    "plate_id": "uuid",
    "body": "Cuts people off constantly.",
    "score": 4,
    "created_at": "2026-05-25T19:00:00Z",
    "author": {
      "id": "uuid",
      "username": "alice"
    }
  }
]
```

| Field | Type | Notes |
|---|---|---|
| `id` | string (UUID) | Review ID — use in subsequent reply/vote requests |
| `plate_id` | string (UUID) | Internal plate record ID |
| `body` | string | Review text |
| `score` | integer | Sum of upvote values; starts at 0 |
| `created_at` | string | UTC timestamp |
| `author.id` | string (UUID) | Internal user ID of the review author |
| `author.username` | string | Display name of the review author |

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | Invalid state or plate parameter |
| 500 | `internal_error` | Unexpected server error |

---

#### `POST /plates/{state}/{plate}/reviews`

Post a new review for a plate.

**Auth required:** yes

**URL parameters** — same as `GET /plates/{state}/{plate}`.

**Request body**
```json
{
  "body": "Tailgates constantly on the highway."
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `body` | string | yes | Whitespace-only strings are rejected |

**Response `201 Created`** — same shape as a single item from `GET /plates/{state}/{plate}/reviews`.

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | Empty body, invalid state/plate, or malformed JSON |
| 401 | `unauthorized` | Missing or invalid token |
| 404 | `not_found` | Authenticated user has no user row |
| 500 | `internal_error` | Unexpected server error |

---

#### `DELETE /reviews/{id}`

Delete a review. Only the review's author may delete it.

**Auth required:** yes

**URL parameters**

| Param | Example |
|---|---|
| `id` | `a1b2c3d4-...` |

**Response `204 No Content`** — empty body.

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | `id` is not a valid UUID |
| 401 | `unauthorized` | Missing or invalid token |
| 403 | `forbidden` | Authenticated user is not the review author |
| 404 | `not_found` | Review does not exist |
| 500 | `internal_error` | Unexpected server error |

---

### Replies

Replies are threaded one level deep. Top-level replies attach directly to a review. Nested replies attach to another reply. Both share the same response shape; `parent_reply_id` is omitted on top-level replies.

---

#### `GET /reviews/{id}/replies`

List top-level replies on a review, sorted oldest-first.

**Auth required:** no

**URL parameters**

| Param | Notes |
|---|---|
| `id` | Review UUID |

**Response `200 OK`** — array:
```json
[
  {
    "id": "uuid",
    "review_id": "uuid",
    "body": "I agree, almost hit me yesterday.",
    "score": 1,
    "created_at": "2026-05-25T19:00:00Z",
    "author": {
      "id": "uuid",
      "username": "bob"
    }
  }
]
```

`parent_reply_id` is absent on top-level replies.

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | `id` is not a valid UUID |
| 500 | `internal_error` | Unexpected server error |

---

#### `POST /reviews/{id}/replies`

Post a top-level reply on a review.

**Auth required:** yes

**URL parameters**

| Param | Notes |
|---|---|
| `id` | Review UUID |

**Request body**
```json
{
  "body": "Confirmed, nearly rear-ended me too."
}
```

**Response `201 Created`** — same shape as an item from `GET /reviews/{id}/replies`.

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | Empty body, invalid UUID, or malformed JSON |
| 401 | `unauthorized` | Missing or invalid token |
| 404 | `not_found` | Review does not exist, or user has no user row |
| 500 | `internal_error` | Unexpected server error |

---

#### `GET /replies/{id}/replies`

List nested replies under a reply, sorted oldest-first.

**Auth required:** no

**URL parameters**

| Param | Notes |
|---|---|
| `id` | Parent reply UUID |

**Response `200 OK`** — array with the same shape as `GET /reviews/{id}/replies`. `parent_reply_id` is present on every item.

```json
[
  {
    "id": "uuid",
    "review_id": "uuid",
    "parent_reply_id": "uuid",
    "body": "+1, same plate got me last week.",
    "score": 0,
    "created_at": "2026-05-25T19:05:00Z",
    "author": {
      "id": "uuid",
      "username": "carol"
    }
  }
]
```

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | `id` is not a valid UUID |
| 500 | `internal_error` | Unexpected server error |

---

#### `POST /replies/{id}/replies`

Post a nested reply under an existing reply.

**Auth required:** yes

**URL parameters**

| Param | Notes |
|---|---|
| `id` | Parent reply UUID |

**Request body**
```json
{
  "body": "That plate is notorious in this area."
}
```

**Response `201 Created`** — reply object with `parent_reply_id` set.

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | Empty body, invalid UUID, or malformed JSON |
| 401 | `unauthorized` | Missing or invalid token |
| 404 | `not_found` | Parent reply does not exist, or user has no user row |
| 500 | `internal_error` | Unexpected server error |

---

#### `DELETE /replies/{id}`

Delete a reply. Only the reply's author may delete it.

**Auth required:** yes

**URL parameters**

| Param | Notes |
|---|---|
| `id` | Reply UUID |

**Response `204 No Content`** — empty body.

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | `id` is not a valid UUID |
| 401 | `unauthorized` | Missing or invalid token |
| 403 | `forbidden` | Authenticated user is not the reply author |
| 404 | `not_found` | Reply does not exist |
| 500 | `internal_error` | Unexpected server error |

---

### Votes

Upvote-only. A user may cast at most one vote per review or reply. Voting and unvoting are separate requests. The current vote total is returned on every review and reply object via the `score` field — there is no separate endpoint to read votes.

---

#### `POST /reviews/{id}/votes`

Upvote a review.

**Auth required:** yes

**URL parameters**

| Param | Notes |
|---|---|
| `id` | Review UUID |

**Request body** — none required.

**Response `201 Created`**
```json
{
  "id": "uuid",
  "target_type": "review",
  "target_id": "uuid",
  "value": 1,
  "created_at": "2026-05-25T19:10:00Z"
}
```

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | `id` is not a valid UUID |
| 401 | `unauthorized` | Missing or invalid token |
| 404 | `not_found` | Review does not exist, or user has no user row |
| 409 | `conflict` | User has already voted on this review |
| 500 | `internal_error` | Unexpected server error |

---

#### `DELETE /reviews/{id}/votes`

Remove the authenticated user's vote from a review.

**Auth required:** yes

**URL parameters**

| Param | Notes |
|---|---|
| `id` | Review UUID |

**Response `204 No Content`** — empty body.

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | `id` is not a valid UUID |
| 401 | `unauthorized` | Missing or invalid token |
| 404 | `not_found` | No vote exists from this user on this review |
| 500 | `internal_error` | Unexpected server error |

---

#### `POST /replies/{id}/votes`

Upvote a reply.

**Auth required:** yes

**URL parameters**

| Param | Notes |
|---|---|
| `id` | Reply UUID |

**Response `201 Created`** — same shape as `POST /reviews/{id}/votes` with `"target_type": "reply"`.

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | `id` is not a valid UUID |
| 401 | `unauthorized` | Missing or invalid token |
| 404 | `not_found` | Reply does not exist, or user has no user row |
| 409 | `conflict` | User has already voted on this reply |
| 500 | `internal_error` | Unexpected server error |

---

#### `DELETE /replies/{id}/votes`

Remove the authenticated user's vote from a reply.

**Auth required:** yes

**URL parameters**

| Param | Notes |
|---|---|
| `id` | Reply UUID |

**Response `204 No Content`** — empty body.

**Error responses**

| Status | Code | When |
|---|---|---|
| 400 | `invalid_input` | `id` is not a valid UUID |
| 401 | `unauthorized` | Missing or invalid token |
| 404 | `not_found` | No vote exists from this user on this reply |
| 500 | `internal_error` | Unexpected server error |

---

## CORS

The server allows `GET`, `POST`, `PATCH`, `DELETE`, and `OPTIONS` from the configured `CORS_ORIGIN`. Credentials (cookies, Authorization headers) are permitted. Send a preflight `OPTIONS` request before cross-origin requests with custom headers.

---

## Typical frontend flows

### New user onboarding

1. User signs up via Clerk.
2. Clerk SDK provides a session token.
3. Frontend calls `POST /users` with the chosen username and the Bearer token.
4. On `409 conflict`: prompt the user to pick a different username and retry.
5. On success: store the returned user object; the user is now registered.

### Looking up a plate

1. User enters a state and plate number.
2. Call `GET /plates/{state}/{plate}/reviews`.
   - A `400` means the state or plate format is invalid — show a validation error.
   - A `200` with an empty array means the plate exists (or is being created) but has no reviews yet.
3. Optionally call `GET /plates/{state}/{plate}` to get the plate's internal `id` for other operations.

### Posting a review

1. User must be authenticated with a valid token and have completed onboarding (`POST /users`).
2. Call `POST /plates/{state}/{plate}/reviews` with `{ "body": "..." }`.
3. On `404`: the user hasn't registered yet — redirect to onboarding.

### Voting

1. Render a vote button with a count from `score` on the review/reply object.
2. On click: call `POST /reviews/{id}/votes` (or `/replies/{id}/votes`).
3. On `409 conflict`: the user already voted — update UI to show voted state without an error message.
4. To undo: call `DELETE /reviews/{id}/votes`.

### Reading replies

Top-level replies: `GET /reviews/{id}/replies`  
Nested replies under a reply: `GET /replies/{id}/replies`

Fetch nested replies lazily (e.g. on "show replies" expand) rather than recursively up front.
