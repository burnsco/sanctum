# Audit Fix Implementation

## Metadata

- Date: `2026-05-05`
- Branch: main
- Author/Agent: Codex
- Scope: Backend auth/session bootstrap, websocket auth, chat abuse controls, uploads, migrations, SQL logging

## Structured Signals

```json
{
  "Report-Version": "1.0",
  "Domains": ["backend", "auth", "websocket", "db", "redis"],
  "Lessons": [
    {
      "title": "Commit audit fixes as independently reviewable units",
      "severity": "MEDIUM",
      "anti_pattern": "Batching unrelated security and reliability fixes into one large commit",
      "detection": "git log --oneline",
      "prevention": "Use one commit per audit finding and run targeted tests before moving to the next item"
    }
  ]
}
```

## Summary

- Requested: work through `docs/reports/2026-05-05-1701-full-stack-code-audit.md` with commits between items and push directly to main.
- Delivered: implemented each listed backend finding in a separate commit, then validated the backend suite.

## Changes Made

- Made Redis initialization return errors and fail closed in production/staging bootstrap paths.
- Bound websocket ticket retry cache to process-local request fingerprints and removed Redis consumed-ticket replay state.
- Enforced group chat participant additions through owner/admin policy and block checks.
- Added HTTP and handler-level image upload size limits before full read/decode.
- Wrapped migration SQL and migration log recording in one transaction.
- Redacted quoted SQL literals before slow/error query logging.

## Validation

- Commands run:
  - `go test ./internal/cache ./internal/bootstrap ./internal/server`
  - `go test ./internal/service`
  - `go test ./internal/database`
  - `make fmt`
  - `make test-backend`
- Test results:
  - Backend race test suite passed.
- Manual verification:
  - Confirmed six implementation commits were created before documentation commit.

## Risks and Regressions

- Known risks:
  - Websocket retry binding uses request metadata available to Fiber; exact same-client replay inside the retry TTL is still best mitigated by immediate `consumeWSTicket` on successful websocket setup.
  - Group chat policy now limits participant adds to owners/admins, which is a behavior tightening.
- Potential regressions:
  - Deployments with Redis unavailable in staging/production now fail startup instead of running degraded.
- Mitigations:
  - Added focused regression tests for each changed surface and ran the full backend suite with race detector.

## Follow-ups

- Remaining work:
  - Frontend tests were not needed for these backend-only fixes.
- Recommended next steps:
  - Monitor staging startup after Redis fail-closed behavior lands.

## Rollback Notes

- Revert the individual commits for the specific audit item that needs rollback.
