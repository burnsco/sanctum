# Task Report

## Metadata

- Date: `2026-04-09`
- Branch: `main`
- Author/Agent: `Codex`
- Scope: `Follow-up fixes for the last two deep audit findings: paginated chat cache keys, realtime test/mock sync, and verification baseline cleanup`

## Structured Signals

```json
{
  "Report-Version": "1.0",
  "Domains": ["backend", "frontend", "websocket", "redis", "docs"],
  "Lessons": [
    {
      "title": "Version paginated room-history caches and include pagination in the cache key",
      "severity": "MEDIUM",
      "anti_pattern": "Caching chat history only by room ID returns stale or incorrect pages when limit/offset vary and makes invalidation too coarse",
      "detection": "rg -n \"MessageHistoryKey\\(|Limit\\(limit\\)|Offset\\(offset\\)\" backend/internal/cache backend/internal/repository/chat.go",
      "prevention": "Use room-scoped cache versioning plus limit/offset in the key so all cached pages stay isolated and invalidation remains cheap"
    },
    {
      "title": "Keep websocket test doubles aligned with EventTarget-based runtime listeners",
      "severity": "MEDIUM",
      "anti_pattern": "Tests continue mocking only onopen/onmessage even after production hooks attach listeners via addEventListener, leaving frontend tests and type-check red",
      "detection": "make test-frontend && cd frontend && bun run type-check && rg -n \"onopen:|onmessage:|addEventListener\\(\" frontend/src -g '*.test.ts' -g '*.test.tsx'",
      "prevention": "Give websocket mocks addEventListener-compatible behavior, keep test guidance current, and update config typing when build-tool contracts change"
    }
  ]
}
```

## Summary

- Requested: fix the last two findings from the deep code audit report.
- Delivered: versioned paginated chat-history cache keys, regression coverage for cache invalidation, synced websocket mocks/tests with `addEventListener`, restored frontend type-check/tests, and updated the stale backend limiter test expectation called out in the audit validation.

## Changes Made

- Updated [inventory.go](/home/cburns/apps/sanctum/backend/internal/cache/inventory.go) so room message history keys now include cache version, `limit`, and `offset`, with room invalidation rotating the message-history version instead of deleting a single unpaginated key.
- Updated [chat.go](/home/cburns/apps/sanctum/backend/internal/repository/chat.go) to use the new paginated cache key.
- Added [inventory_test.go](/home/cburns/apps/sanctum/backend/internal/cache/inventory_test.go) to verify page isolation and cache-key rotation after room invalidation.
- Updated [middleware_cors_test.go](/home/cburns/apps/sanctum/backend/internal/server/middleware_cors_test.go) so the limiter assertions match the current 300-requests-per-minute middleware configuration.
- Fixed the `useManagedWebSocket` rate-limit narrowing in [useManagedWebSocket.ts](/home/cburns/apps/sanctum/frontend/src/hooks/useManagedWebSocket.ts) and replaced object-form Rollup chunk configuration with a typed function in [vite.config.ts](/home/cburns/apps/sanctum/frontend/vite.config.ts).
- Updated websocket test doubles in [useManagedWebSocket.test.ts](/home/cburns/apps/sanctum/frontend/src/hooks/useManagedWebSocket.test.ts), [useGameRoomSession.test.tsx](/home/cburns/apps/sanctum/frontend/src/hooks/useGameRoomSession.test.tsx), [ChatProvider.spec.tsx](/home/cburns/apps/sanctum/frontend/src/providers/ChatProvider.spec.tsx), and [ChatProvider.integration.test.tsx](/home/cburns/apps/sanctum/frontend/src/providers/ChatProvider.integration.test.tsx) to support `addEventListener`-driven runtime behavior.
- Updated [TESTING.md](/home/cburns/apps/sanctum/frontend/TESTING.md) so the documented websocket mocking pattern matches the current hook contract.

## Validation

- Commands run:
  - `gofmt -w backend/internal/cache/inventory.go backend/internal/cache/inventory_test.go backend/internal/repository/chat.go backend/internal/server/middleware_cors_test.go`
  - `cd backend && go test ./internal/cache -count=1`
  - `cd backend && go test ./internal/server -run 'TestSetupMiddleware_(RateLimitedResponseIncludesCORSHeaders|PreflightBypassesLimiter)' -count=1`
  - `cd backend && go test ./internal/repository -count=1`
  - `cd frontend && bun run type-check`
  - `cd frontend && bun run vitest run src/hooks/useManagedWebSocket.test.ts src/hooks/useGameRoomSession.test.tsx src/providers/ChatProvider.spec.tsx src/providers/ChatProvider.integration.test.tsx`
  - `make test-backend`
  - `make test-frontend`
  - `make lint-frontend`
  - `make lint`
- Test results:
  - Focused backend cache/server/repository tests passed.
  - Focused realtime/frontend websocket suites passed.
  - `make test-backend` passed.
  - `make test-frontend` passed.
  - `make lint-frontend` passed with existing warnings only.
  - `make lint` still fails on the pre-existing `gosec` findings in `backend/internal/service/image_service.go`.
- Manual verification:
  - Not performed in a running browser session.

## Risks and Regressions

- Known risks:
  - Versioned message-history keys can temporarily increase Redis key cardinality until the old page entries age out by TTL.
- Potential regressions:
  - New websocket mocks now emit the initial `"connected"` handshake message, so future tests that assert first-call payloads need to account for that runtime behavior.
- Mitigations:
  - Added explicit cache-key regression coverage.
  - Updated websocket test helpers and test guidance together so the contract stays documented.

## Follow-ups

- Remaining work:
  - Address the existing `gosec` findings in `backend/internal/service/image_service.go` so `make lint` can go green.
  - Clean up the existing frontend lint warnings if the team wants a warning-free `oxlint` baseline.
- Recommended next steps:
  - If chat-history caching is expanded again, prefer cursor-aware or versioned keys over room-only cache entries.
  - Reuse the updated websocket mock pattern for any new realtime hook/provider tests.

## Rollback Notes

- Revert the paginated message-history key/versioning changes in [inventory.go](/home/cburns/apps/sanctum/backend/internal/cache/inventory.go), [inventory_test.go](/home/cburns/apps/sanctum/backend/internal/cache/inventory_test.go), and [chat.go](/home/cburns/apps/sanctum/backend/internal/repository/chat.go) to restore the old room-only cache behavior.
- Revert the websocket mock/test updates plus [vite.config.ts](/home/cburns/apps/sanctum/frontend/vite.config.ts) and [useManagedWebSocket.ts](/home/cburns/apps/sanctum/frontend/src/hooks/useManagedWebSocket.ts) if you need to return to the previous frontend test/build setup.
