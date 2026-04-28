# Local Review Fixes

## Metadata

- Date: `2026-04-28`
- Branch: `codex/local-review-fixes`
- Author/Agent: Codex
- Scope: Backend security/realtime/visibility hardening, frontend auth storage, tests, and ops cleanup

## Structured Signals

```json
{
  "Report-Version": "1.0",
  "Domains": ["backend", "frontend", "auth", "websocket", "db", "redis", "infra"],
  "Lessons": [
    {
      "title": "Review reports should be converted into test-backed fixes before publishing",
      "severity": "HIGH",
      "anti_pattern": "Leaving local review findings as unchecked documentation lets security and realtime regressions persist without executable coverage",
      "detection": "rg -n \"\\[ \\].*HIGH|BroadcastToConversation|auth-session-storage|visibility\" docs/reports backend frontend",
      "prevention": "Pair each actionable review item with a bounded code change and focused validation before opening the PR"
    }
  ]
}
```

## Summary

- Implemented fixes from the local full-stack review report.
- Preserved existing API behavior where possible while tightening public data exposure, auth storage, chat delivery, and operational shutdown paths.

## Changes Made

- Removed duplicate HTTP chat WebSocket delivery and added Redis-publish coverage for one chat fanout per HTTP send.
- Added explicit conversation visibility with SQL migration and public-chatroom gating.
- Replaced fragile public user mutation with safe summary DTOs for public user endpoints and stricter sanitizer behavior.
- Moved the frontend access token to memory-only state and bootstraps protected routes through the HttpOnly refresh cookie.
- Added retry-loop protection for 401 refresh handling.
- Hardened logout refresh-token revocation, JTI generation, refresh-token Redis context usage, admin cache behavior, metrics auth, proxy warnings, and shutdown deadlines.
- Deduplicated server constructor initialization and removed backend scratch files.
- Addressed PR review follow-ups:
  - Made `000013_conversation_visibility` skip adding the visibility check constraint when the baseline schema already created it.
  - Ensured logout clears the refresh cookie before access-token blacklist or refresh-token revocation errors can return.
  - Made admin cache read/write failures fall back to the database result instead of blocking admin authorization.

## Validation

- Commands run:
  - `make fmt`
  - `make fmt-frontend`
  - `cd frontend && bun run type-check`
  - `go test ./internal/server ./internal/service`
  - `cd frontend && bun run test:run src/stores/useAuthSessionStore.test.ts src/hooks/useManagedWebSocket.test.ts`
  - `make test-backend`
  - `make lint`
  - `make lint-frontend`
  - `make test-frontend`
  - `git diff --check`
  - `go test ./internal/server ./internal/service`
- Test results:
  - Backend race suite passed.
  - Frontend suite passed: 60 files, 267 tests.
  - Backend and frontend lint passed.
  - PR follow-up backend targeted suite passed.
- Manual verification:
  - Reviewed final diff/stat and confirmed changes map to the local review findings.

## Risks and Regressions

- Known risks:
  - Public user responses now intentionally omit fields such as email and timestamps in public contexts.
  - Existing browser sessions relying on persisted access tokens will rely on refresh-cookie bootstrap after deployment.
- Potential regressions:
  - Deployments must apply migration `000013_conversation_visibility` before using visibility-gated chatroom queries.
- Mitigations:
  - Added migration defaults for existing rows and updated frontend types to tolerate public user summaries.

## Follow-ups

- Remaining work:
  - Run OpenAPI generation/checks if API documentation must reflect public user summary response shapes before merge.
- Recommended next steps:
  - Review the PR with attention to auth bootstrap UX and chatroom migration rollout.

## Rollback Notes

- Revert the PR to restore previous behavior.
- If rolling back only the DB change, apply `000013_conversation_visibility.down.sql` after dependent code is reverted.
