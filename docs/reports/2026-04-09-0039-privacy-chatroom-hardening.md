# Task Report

## Metadata

- Date: `2026-04-09`
- Branch: `main`
- Author/Agent: `Codex`
- Scope: `Backend privacy hardening for shared user payloads and public chatroom exposure`

## Structured Signals

```json
{
  "Report-Version": "1.0",
  "Domains": ["backend", "auth", "websocket", "redis"],
  "Lessons": [
    {
      "title": "Do not serialize shared GORM user models directly on public or multi-user surfaces",
      "severity": "HIGH",
      "anti_pattern": "Returning raw models.User values through public profiles, nested post/comment payloads, friends, chat, and realtime responses leaks private fields such as email and moderation state",
      "detection": "rg -n \"c\\.JSON\\(|BroadcastToConversation\\(|publishBroadcastEvent\\(\" backend/internal/server",
      "prevention": "Sanitize shared user payloads at server response boundaries and keep full user data only on self/admin routes"
    },
    {
      "title": "Treat only sanctum-backed group conversations as public chatrooms",
      "severity": "HIGH",
      "anti_pattern": "Using is_group alone for chatroom discovery/joining makes user-created private groups globally discoverable and joinable",
      "detection": "rg -n \"GetAllChatrooms|GetJoinedChatrooms|JoinChatroom|sanctum_id\" backend/internal/service backend/internal/server",
      "prevention": "Require SanctumID on public chatroom queries and reject non-public rooms from the /chatrooms surface"
    }
  ]
}
```

## Summary

- Requested: fix the two high-severity backend findings from the deep audit.
- Delivered: shared-response user sanitization across public/multi-user handlers, stricter public chatroom filtering and join rules, cache key rotation for chatroom discovery, and regression tests.

## Changes Made

- Added [response_sanitizers.go](/home/cburns/apps/sanctum/backend/internal/server/response_sanitizers.go) to scrub `email`, moderation metadata, and other private user fields from shared payloads.
- Applied sanitization to user, post, comment, friend, chat, mention, block, chat moderation, game room, cached-user, and personal sanctum-request handlers.
- Restricted chatroom listing and joining in [chat_service.go](/home/cburns/apps/sanctum/backend/internal/service/chat_service.go) to `is_group = true` conversations with `sanctum_id IS NOT NULL`.
- Rotated the public chatroom list cache namespace in [inventory.go](/home/cburns/apps/sanctum/backend/internal/cache/inventory.go) so stale Redis entries do not keep exposing old room lists after deploy.
- Added regression coverage in [user_handlers_test.go](/home/cburns/apps/sanctum/backend/internal/server/user_handlers_test.go), [response_sanitizers_test.go](/home/cburns/apps/sanctum/backend/internal/server/response_sanitizers_test.go), and [chat_service_test.go](/home/cburns/apps/sanctum/backend/internal/service/chat_service_test.go).

## Validation

- Commands run:
  - `make fmt`
  - `cd backend && go test ./internal/server -run 'Test(GetUserProfile|GetMyProfile|SanitizeShared|GetMyBlocks|GetMyMentions)' -count=1 -v`
  - `cd backend && go test ./internal/service -run 'TestChatService' -count=1 -v`
  - `cd backend && go test ./internal/cache -count=1`
  - `make test-backend`
  - `make lint`
- Test results:
  - Focused server/service/cache tests passed.
  - `make test-backend` still fails on the pre-existing limiter expectation in `backend/internal/server/middleware_cors_test.go`.
  - `make lint` still fails on pre-existing `gosec` findings in `backend/internal/service/image_service.go`.
- Manual verification:
  - Not performed in a running browser session.

## Risks and Regressions

- Known risks:
  - Some frontend screens that previously showed other users' emails may now render empty strings until the UI is adjusted.
- Potential regressions:
  - Any handler still returning raw nested `models.User` outside the patched surfaces could remain exposed.
- Mitigations:
  - Added shared sanitizer tests and route/service regression coverage on the highest-risk paths.

## Follow-ups

- Remaining work:
  - Patch frontend views to hide or replace blank email rows for non-self users.
  - Fix the chat message pagination cache key issue from the original audit.
  - Clean up the pre-existing backend test/lint baseline so full validation can go green.
- Recommended next steps:
  - Update public profile and friend UI to stop assuming `email` is always present.
  - Audit any remaining admin-adjacent shared payloads for deliberate versus accidental data exposure.

## Rollback Notes

- Revert the sanitizer wiring plus [response_sanitizers.go](/home/cburns/apps/sanctum/backend/internal/server/response_sanitizers.go) to restore prior response payloads.
- Revert the `sanctum_id` chatroom query restrictions and cache key rotation if public chatroom behavior must temporarily return to the old model.
