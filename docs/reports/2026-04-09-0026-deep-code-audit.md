# Task Report Template

## Metadata

- Date: `2026-04-09`
- Branch: `main`
- Author/Agent: `Codex`
- Scope: `Deep code audit of backend/frontend security, access control, realtime, and verification health`

## Structured Signals

```json
{
  "Report-Version": "1.0",
  "Domains": ["backend", "frontend", "auth", "websocket", "docs"],
  "Lessons": [
    {
      "title": "Do not serialize raw User models to non-admin clients",
      "severity": "HIGH",
      "anti_pattern": "Shared User structs expose email and moderation state across posts, comments, chat, friends, and user-list endpoints",
      "detection": "rg -n \"json:\\\"email|json:\\\"is_banned|json:\\\"banned_reason|Preload\\(\\\"User\\\"\\)|Preload\\(\\\"Participants\\\"\\)|Preload\\(\\\"Sender\\\"\\)\" backend/internal",
      "prevention": "Use explicit public DTOs for user summaries and reserve sensitive fields for self/admin-only responses"
    },
    {
      "title": "Do not treat every group conversation as a public chatroom",
      "severity": "HIGH",
      "anti_pattern": "User-created group conversations can be discovered and joined through the chatroom APIs because visibility is inferred only from is_group",
      "detection": "rg -n \"Where\\(\\\"is_group = \\?\\\", true\\)|JoinChatroom|CreateConversation\\(|chatrooms := protected.Group\" backend/internal/service/chat_service.go backend/internal/server/server.go",
      "prevention": "Add an explicit visibility/access model and enforce join/list authorization on the server"
    },
    {
      "title": "Cache paginated message history with pagination in the key",
      "severity": "MEDIUM",
      "anti_pattern": "Message history cache key is only room-scoped while query varies by limit/offset",
      "detection": "rg -n \"MessageHistoryKey\\(|Limit\\(limit\\)|Offset\\(offset\\)\" backend/internal/cache backend/internal/repository/chat.go",
      "prevention": "Include limit/offset or cursor identity in cache keys, or cache only canonical unpaginated slices"
    },
    {
      "title": "Keep realtime code, mocks, and TS config in sync",
      "severity": "MEDIUM",
      "anti_pattern": "Runtime code now expects EventTarget-style websockets while tests still provide onopen/onmessage-only mocks, and frontend type-check currently fails",
      "detection": "make test-frontend && cd frontend && bun run type-check",
      "prevention": "Update test doubles whenever the socket contract changes and keep type-check green in CI"
    }
  ]
}
```

## Summary

- Requested a deep audit of the site.
- Delivered a review of backend/frontend risk areas with concrete findings, validation results, and follow-up priorities.

## Changes Made

- Added this audit report only.
- No product code was changed during the audit.

## Validation

- Commands run:
  - `make test-backend`
  - `make test-frontend`
  - `cd frontend && bun run type-check`
- Test results:
  - `make test-backend` failed in `backend/internal/server/middleware_cors_test.go` because the test still expects POSTs to rate-limit after 100 requests while middleware is configured for 300 requests/minute outside dev/test/stress.
  - `make test-frontend` failed across websocket-related suites because `useManagedWebSocket` now attaches listeners with `addEventListener`, but test mocks still expose only `onopen`/`onmessage`-style handlers.
  - `cd frontend && bun run type-check` failed on `frontend/src/hooks/useManagedWebSocket.ts` and `frontend/vite.config.ts`.
- Manual verification:
  - Traced auth/session, user/profile, posts/comments, friends, sanctums, chat/chatroom, and websocket paths in source.

## Risks and Regressions

- Known risks:
  - Sensitive user fields are over-exposed by shared response models.
  - Group chat visibility/joins are under-authorized.
  - Chat history pagination can return stale or wrong pages from cache.
  - Verification baseline is red, especially around realtime/frontend code.
- Potential regressions:
  - Tightening user DTOs will require frontend type updates.
  - Adding chatroom visibility rules will affect existing group conversation behavior and tests.
  - Fixing message-history cache keys may change perceived pagination behavior and cache hit rate.
- Mitigations:
  - Introduce explicit DTO layers.
  - Add regression tests for unauthorized room listing/joining.
  - Add cache-key tests for paginated history.
  - Repair type-check and websocket mocks before further realtime refactors.

## Follow-ups

- Remaining work:
  - Replace raw `models.User` serialization with public/self/admin DTOs.
  - Add explicit public/private semantics for group conversations and gate `GetAllChatrooms`/`JoinChatroom`.
  - Fix paginated message-history caching.
  - Restore green frontend type-check/tests and update stale backend limiter test expectations.
- Recommended next steps:
  - Prioritize the two access/privacy issues before feature work.
  - Add dedicated security tests for user-field exposure and unauthorized chatroom discovery/join.

## Rollback Notes

- Revert this report by deleting `docs/reports/2026-04-09-0026-deep-code-audit.md` if needed.
