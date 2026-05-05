# Full Stack Code Audit

## Metadata

- Date: `2026-05-05`
- Branch: main
- Author/Agent: Codex
- Scope: Backend, frontend, auth/session, websocket, database, media upload, logging

## Structured Signals

```json
{
  "Report-Version": "1.0",
  "Domains": ["backend", "frontend", "auth", "websocket", "db", "redis", "infra"],
  "Lessons": [
    {
      "title": "Fail closed when Redis-backed auth state is unavailable",
      "severity": "HIGH",
      "anti_pattern": "Continuing runtime without Redis while refresh-token rotation, token revocation, websocket tickets, and rate limits depend on Redis",
      "detection": "rg -n \"continuing without cache|s\\.redis != nil|refresh_token:\" backend/internal",
      "prevention": "Make Redis availability required in prod-like environments or add a durable fallback for refresh-token state and revocation"
    },
    {
      "title": "Keep websocket tickets strictly single use",
      "severity": "HIGH",
      "anti_pattern": "A consumed websocket ticket is accepted from a short-lived cache for handshake retries without binding it to the original upgrade attempt",
      "detection": "rg -n \"consumedTickets|ws_ticket_consumed|GetDel\" backend/internal/server",
      "prevention": "Replace reusable consumed-ticket cache with handshake-specific state or delete consumed cache immediately after the known Fiber retry path"
    },
    {
      "title": "Validate social graph writes at the service boundary",
      "severity": "MEDIUM",
      "anti_pattern": "Group conversations can add arbitrary participant IDs without consent, relationship, or block checks",
      "detection": "rg -n \"AddParticipant|ParticipantIDs|usersBlocked\" backend/internal/service/chat_service.go",
      "prevention": "Require invite acceptance or enforce friendship/block checks for all conversation participant writes"
    }
  ]
}
```

## Summary

Requested: deep audit of the full-stack social media app.

Delivered: prioritized code-review findings across auth/session, websocket, chat abuse controls, uploads, migrations, and observability. No code changes were made outside this report.

## Findings

### High: Redis can disappear while auth semantics silently degrade

`backend/internal/cache/redis.go` continues with `client = nil` on invalid URL or failed ping. Several auth paths then branch on `s.redis != nil`: refresh-token replay protection only runs when Redis exists, access-token revocation checks only run when Redis exists, and refresh-token issuance skips durable JTI storage when Redis is nil.

Key references:

- `backend/internal/cache/redis.go:44`
- `backend/internal/cache/redis.go:66`
- `backend/internal/server/auth_handlers.go:287`
- `backend/internal/server/auth_handlers.go:467`
- `backend/internal/server/server.go:661`

Impact: in prod-like environments, a Redis outage at startup can turn refresh tokens into reusable bearer credentials until JWT expiry and make logout/revocation ineffective. Some rate-limited routes fail closed, but `/auth/refresh` itself has no route rate limit and no replay check without Redis.

Recommended fix: make Redis connection failure fatal in production/staging, or persist refresh token JTIs/revocation in PostgreSQL with Redis only as cache. Add a test proving `/auth/refresh` rejects reused tokens and fails safely when Redis is unavailable in production mode.

### High: Websocket single-use tickets are reusable during the consumed-ticket cache window

`AuthRequired` atomically `GETDEL`s a websocket ticket, then stores the consumed ticket in process and Redis for 10 seconds so Fiber's multi-pass websocket upgrade can succeed. Any request presenting the same ticket during that cache window is accepted as the original user until `consumeWSTicket` runs in the websocket handler.

Key references:

- `backend/internal/server/server.go:568`
- `backend/internal/server/server.go:690`
- `backend/internal/server/server.go:744`
- `backend/internal/server/server.go:766`
- `frontend/src/lib/ws-utils.ts:43`

Impact: a leaked websocket URL can be replayed briefly, and under race conditions can establish extra authenticated websocket connections. The frontend avoids logging the ticket, which is good, but URLs can still appear in browser/network/proxy surfaces.

Recommended fix: bind the consumed-ticket retry to a handshake nonce/connection fingerprint if possible, keep the compatibility cache process-local only, and add a regression test that two independent websocket upgrade attempts with one ticket cannot both authenticate.

### Medium: Group chat participant writes can force-add users

Direct messages check the block graph before creating or reusing a one-on-one conversation. Group conversation creation and `AddParticipant` do not apply the same block/consent checks, and any existing group participant can add arbitrary user IDs.

Key references:

- `backend/internal/service/chat_service.go:81`
- `backend/internal/service/chat_service.go:136`
- `backend/internal/service/chat_service.go:235`

Impact: blocked users can be reintroduced through group chats, and users can be pulled into unwanted conversations. For a social app, this is an abuse-control gap even if message sending still validates membership.

Recommended fix: require invite/acceptance for non-public groups, or at least enforce block checks and friend/admin policy before adding participants. Add tests for "blocked user cannot be added to group" and "non-owner participant cannot add arbitrary users" if that is the intended policy.

### Medium: Image upload size enforcement happens after full read/decode

The handler reads the entire uploaded file into memory with `io.ReadAll`, then `ImageService.Upload` checks `len(in.Content)` against `IMAGE_MAX_UPLOAD_SIZE_MB`. The Fiber app does not set `BodyLimit`, so the configured app-level image limit is not clearly enforced at the transport boundary.

Key references:

- `backend/internal/server/image_handlers.go:49`
- `backend/internal/service/image_service.go:129`
- `backend/internal/server/server.go:836`

Impact: this can cause confusing rejects below the configured 10MB limit if Fiber's default body limit wins, or memory pressure if proxy/server body limits are raised without streaming checks.

Recommended fix: set `fiber.Config.BodyLimit` from `ImageMaxUploadSizeMB` plus multipart overhead, reject by `file.Size` before reading, and cap reads with `io.LimitReader`.

### Medium: Migrations are not recorded atomically with their SQL

`ApplyMigration` executes the migration SQL and then inserts the migration log in a separate operation. If SQL succeeds and the log insert fails, the schema can be changed without a corresponding `migration_logs` row.

Key references:

- `backend/internal/database/migrate_runner.go:60`
- `backend/internal/database/migrate_runner.go:69`

Impact: future deploys can retry already-applied DDL, causing startup failures or requiring manual repair.

Recommended fix: wrap migration SQL plus log insert in a transaction where PostgreSQL permits it. For non-transactional operations, mark migrations explicitly and document/manual-guard them.

### Low: Slow/error SQL logging can include sensitive values

The custom GORM logger writes the full rendered SQL into error and slow-query logs. Depending on GORM interpolation and the query, this can expose emails, profile content, password hashes, moderation content, or tokens if future queries include them.

Key references:

- `backend/internal/database/database.go:70`
- `backend/internal/database/database.go:75`
- `backend/internal/database/database.go:82`

Recommended fix: configure parameterized SQL logging without values for production, or redact sensitive table/column patterns before log emission.

## Positive Notes

- Backend layering is consistent: handlers generally delegate to services and repositories.
- Public response sanitizers are present and used broadly for embedded users.
- Backend test coverage is meaningful across handlers, repositories, services, websocket hubs, and validation.
- Access tokens are memory-only in the frontend store, while refresh tokens are HTTP-only cookies.
- Websocket clients request tickets rather than sending JWTs in query params.
- Production config validation catches weak JWT secrets, default DB password, disabled DB SSL, and missing Redis URL.

## Validation

- Commands run:
  - `git status --short`
  - `make test-backend`
  - Static source audit with `rg`, `find`, `sed`, and `nl`
- Test results:
  - `make test-backend` passed with race detector across backend packages.
- Not run:
  - Frontend tests/type-check. `frontend/node_modules` was missing, and this audit did not install dependencies.

## Risks and Regressions

- This was a static/manual audit, not a penetration test.
- I did not run the Docker Compose stack or exercise browser flows.
- Dependency vulnerability freshness was not checked against live advisories.

## Follow-ups

- Fix the Redis/auth fail-closed behavior first.
- Add websocket ticket replay regression coverage.
- Decide the intended group-chat invite/consent policy, then enforce it service-side.
- Wire upload limits at the HTTP boundary.
- Make migration logging transactional.

## Rollback Notes

This report is documentation-only. Remove `docs/reports/2026-05-05-1701-full-stack-code-audit.md` to revert it.
