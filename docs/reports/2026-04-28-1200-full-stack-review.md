# Full-Stack Review — 2026-04-28

## Metadata

- Date: `2026-04-28`
- Branch: `main`
- Author/Agent: Claude Sonnet 4.6
- Scope: Full-stack security, architecture, realtime, tests, and ops review

## Structured Signals

```json
{
  "Report-Version": "1.0",
  "Domains": ["backend", "frontend", "auth", "websocket", "redis", "infra"],
  "Lessons": [
    {
      "title": "HTTP SendMessage calls BroadcastToConversation AND PublishChatMessage — double delivery",
      "severity": "HIGH",
      "anti_pattern": "chat_handlers.go calls chatHub.BroadcastToConversation directly then notifier.PublishChatMessage; StartWiring re-broadcasts via Redis, so local WS clients receive the message twice",
      "detection": "rg -n \"BroadcastToConversation|PublishChatMessage\" backend/internal/server/chat_handlers.go",
      "prevention": "HTTP SendMessage should only publish to Redis (like the WS handler does); remove the direct BroadcastToConversation call from the HTTP path"
    },
    {
      "title": "Mutation-based response sanitization is fragile — new User fields silently escape",
      "severity": "HIGH",
      "anti_pattern": "sanitizeSharedUser() zeroes sensitive fields on the live struct; any new field added to models.User escapes all public endpoints until manually added to the sanitizer",
      "detection": "rg -n \"sanitizeSharedUser\\|json:\\\"email\" backend/internal/server/response_sanitizers.go",
      "prevention": "Introduce a public UserSummary DTO {ID, Username, Avatar} and marshal that instead of the raw model"
    },
    {
      "title": "Group conversation and chatroom visibility share only is_group flag",
      "severity": "HIGH",
      "anti_pattern": "User-created group DMs and sanctum chatrooms both have is_group=true; GetAllChatrooms gates on that flag, so user DMs can surface via chatroom APIs",
      "detection": "rg -n \"is_group = \\?.*true|GetAllChatrooms|JoinChatroom\" backend/internal/service/chat_service.go backend/internal/server/server.go",
      "prevention": "Add explicit visibility enum to Conversation (public/private/direct) and gate chatroom list/join on it"
    },
    {
      "title": "Access token stored in localStorage is XSS-stealable",
      "severity": "HIGH",
      "anti_pattern": "useAuthSessionStore persists accessToken to localStorage via Zustand; XSS can exfiltrate it",
      "detection": "rg -n \"auth-session-storage|accessToken\" frontend/src/stores/useAuthSessionStore.ts",
      "prevention": "Keep access token in memory only; rely on the HttpOnly refresh cookie for persistence and bootstrap /users/me on load"
    },
    {
      "title": "Frontend test suite and type-check are broken",
      "severity": "HIGH",
      "anti_pattern": "useManagedWebSocket switched to addEventListener but mocks still use onopen/onmessage; vite.config.ts has type errors; CORS test expects old 100-req limit",
      "detection": "make test-frontend && cd frontend && bun run type-check",
      "prevention": "Update mocks whenever the WebSocket contract changes; keep type-check green in CI"
    },
    {
      "title": "Logout does not validate Redis Del result",
      "severity": "MEDIUM",
      "anti_pattern": "auth_handlers.go ignores s.redis.Del return value at logout; if Del fails the refresh token remains valid while the user believes they are logged out",
      "detection": "rg -n \"redis.Del\" backend/internal/server/auth_handlers.go",
      "prevention": "Log (or return) Del errors at logout, consistent with the fail-closed pattern used in Refresh"
    },
    {
      "title": "/metrics and /api/metrics/dashboard are unauthenticated",
      "severity": "MEDIUM",
      "anti_pattern": "Prometheus scrape endpoint and Fiber monitor dashboard are publicly reachable without auth",
      "detection": "rg -n \"RegisterAt|metrics/dashboard\" backend/internal/server/server.go",
      "prevention": "Network-isolate or gate behind AdminRequired; at minimum document that nginx must block external access"
    }
  ]
}
```

## Summary

- Requested a full-stack code review of the Sanctum social media application.
- Delivered a structured review covering security, architecture, realtime, testing, and ops concerns.
- No product code was changed; this is a read-only audit.

## Changes Made

- Added this report only.
- Follow-up implementation completed in `docs/reports/2026-04-28-1814-local-review-fixes.md`.

## Validation

- Commands run: read-only source analysis
- Manual verification: traced auth, post, chat, WS, image, admin, and moderation paths in source

## Action Items

Work through these top-to-bottom. Check off each item when done.

### Security — Fix First

- [x] **[HIGH] Fix double delivery on HTTP `SendMessage`**
  - `backend/internal/server/chat_handlers.go:154-166`
  - Remove the direct `chatHub.BroadcastToConversation` calls from the HTTP handler.
  - Let the existing `notifier.PublishChatMessage` → Redis → `StartWiring` → `BroadcastToConversation` path handle delivery (same as the WS handler does).
  - Add a test asserting exactly one broadcast per HTTP message send.

- [x] **[HIGH] Replace mutation-based sanitizer with a UserSummary DTO**
  - `backend/internal/server/response_sanitizers.go`
  - Define `type UserSummary struct { ID uint; Username string; Avatar string }` in models or a dto package.
  - Replace all public endpoints that currently embed `models.User` with `UserSummary`.
  - Keep the raw `models.User` in self-profile (`/users/me`) and admin endpoints only.

- [x] **[HIGH] Add explicit visibility to Conversation model**
  - `backend/internal/models/chat.go`
  - Add `Visibility string` enum (`public` / `private` / `direct`) with a migration.
  - Gate `GetAllChatrooms` and `JoinChatroom` on `visibility = 'public'`.
  - Mark user-created group conversations as `private` on creation.

- [x] **[HIGH] Move access token out of localStorage**
  - `frontend/src/stores/useAuthSessionStore.ts`
  - Remove `partialize` so the access token is not persisted.
  - On app load, attempt token refresh from the HttpOnly cookie to bootstrap auth state.
  - Remove the redundant `localStorage.setItem("user", ...)` calls in `useAuth.ts`.

- [x] **[MEDIUM] Validate Redis `Del` result in `Logout`**
  - `backend/internal/server/auth_handlers.go` Logout handler
  - Log the error if `Del` fails; optionally return 500 instead of 200.

- [x] **[MEDIUM] Gate `/metrics` and `/api/metrics/dashboard`**
  - `backend/internal/server/server.go:371-374`
  - Either move behind `AdminRequired()` or document the nginx firewall requirement explicitly.

- [x] **[MEDIUM] Fix 401 retry loop in API client**
  - `frontend/src/api/client.ts` `request()` method
  - Add a `retried` flag to prevent a second refresh attempt on the retry request.

- [x] **[LOW] Use full UUID for JTI**
  - `backend/internal/server/auth_handlers.go:generateJTI()`
  - Replace `uuid.New().String()[:8]` with the full `uuid.New().String()`.

- [x] **[LOW] Pass request context to `generateRefreshToken` Redis Set**
  - `backend/internal/server/auth_handlers.go:465`
  - Change `context.Background()` to the caller's `c.Context()`.

---

### Architecture & Code Quality

- [x] **[HIGH] Deduplicate `NewServer` / `NewServerWithDeps`**
  - `backend/internal/server/server.go:103-199` and `201-287`
  - Extract a shared `initFromDeps(cfg, db, redis)` helper called by both.
  - Risk: the two constructors have diverged in moderation setup; the refactor must verify parity.

- [x] **[MEDIUM] Cache `isAdmin` result in Redis**
  - `backend/internal/server/helpers.go:isAdminByUserID`
  - Add a short TTL (30s) Redis cache keyed on `admin:<userID>`.
  - Invalidate on ban/promote/demote.

- [x] **[LOW] Remove scratch files from backend root**
  - `backend/test_webp.go`, `backend/test_output.txt`, `backend/test_sanctum_debug.txt`
  - Delete or move to `test/` and add to `.gitignore`.

---

### Frontend

- [x] **[MEDIUM] Remove stale `localStorage.setItem("user", ...)` after profile updates**
  - `frontend/src/hooks/useAuth.ts`
  - Once auth is memory-only (see token item above), remove all localStorage user cache writes.

---

### Testing — Restore Green Suite

- [x] **[HIGH] Fix `useManagedWebSocket` test mocks**
  - `frontend/src/hooks/useManagedWebSocket.test.ts`
  - Update mocks to use `addEventListener`/`removeEventListener` interface.

- [x] **[HIGH] Fix frontend type-check failures**
  - `frontend/vite.config.ts` and `frontend/src/hooks/useManagedWebSocket.ts`
  - Run `cd frontend && bun run type-check` and resolve all errors.

- [x] **[MEDIUM] Update CORS middleware test rate-limit expectation**
  - `backend/internal/server/middleware_cors_test.go`
  - Update assertion from 100-request limit to 300-request limit to match current config.

---

### Operations

- [x] **[MEDIUM] Give hub shutdown its own deadline**
  - `backend/cmd/server/main.go:105-115`
  - `srv.Shutdown(ctx)` runs after `app.ShutdownWithContext(ctx)` may have already expired the 10s context. Give `srv.Shutdown` a fresh context with its own timeout.

- [x] **[LOW] Confirm `EnableProxyHeader` is never enabled without a trusted proxy**
  - `backend/cmd/server/main.go:89`
  - Add a startup warning log if `EnableProxyHeader=true` and `Env=production` without a documented proxy setup.

## Risks and Regressions

- Fixing the double-delivery bug changes observable behavior for any client that currently deduplicates by ID — verify client-side dedup logic before deploying.
- Introducing `UserSummary` DTO will require frontend type updates across all hooks and pages that consume `User` objects.
- Adding `visibility` to `Conversation` requires a migration; make it nullable with a default of `public` to avoid breaking existing rows.
- Moving access token out of localStorage will log users out on the next deploy if the bootstrap-from-cookie path isn't ready first.

## Follow-ups

- Implement UserSummary DTO as the core refactor; everything downstream of it improves.
- Get CI green (frontend type-check + tests + backend CORS test) as a forcing function before new realtime features.
- Consider a dedicated security test file that asserts no email field escapes any non-admin/non-self endpoint.

## Rollback Notes

- All action items are independent; each can be reverted individually.
- The Conversation visibility migration must include a rollback: `ALTER TABLE conversations DROP COLUMN visibility`.
