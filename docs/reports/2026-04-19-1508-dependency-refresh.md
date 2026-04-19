# Dependency Refresh Report

## Metadata

- Date: `2026-04-19`
- Branch: `main`
- Author/Agent: `Codex`
- Scope: Frontend and backend dependency refresh plus validation

## Structured Signals

```json
{
  "Report-Version": "1.0",
  "Domains": ["frontend", "backend"],
  "Lessons": [
    {
      "title": "Backend compose-backed tests need Docker availability",
      "severity": "MEDIUM",
      "anti_pattern": "Relying on compose-backed test targets in an environment without a docker CLI causes validation to fail before tests start.",
      "detection": "make test-backend-integration",
      "prevention": "Check Docker availability before selecting compose-backed validation targets, or provide a documented host-only fallback path."
    }
  ]
}
```

## Summary

- Refreshed frontend dependencies to the latest available versions in the current graph, including the remaining `oxfmt` bump to `0.45.0`.
- Refreshed backend Go modules using `go get -u all` and `go mod tidy`, which updated the imported module graph to the latest available releases that the resolver accepted.

## Changes Made

- Updated `frontend/package.json` and `frontend/bun.lock`.
- Updated `backend/go.mod` and `backend/go.sum`.
- No application source code changes were required.

## Validation

- Commands run:
  - `make fmt`
  - `make fmt-frontend`
  - `make lint-frontend`
  - `cd frontend && bun run type-check`
  - `make test-frontend`
  - `make test-backend`
  - `make test-backend-integration`
- Test results:
  - Frontend lint passed with existing warnings.
  - Frontend type-check passed.
  - Frontend Vitest suite passed: `60` files, `267` tests.
  - Backend Go test suite passed.
  - Backend integration target could not start because `docker` is unavailable in this environment.
- Manual verification:
  - `bun outdated` in `frontend/` returned no remaining outdated packages.

## Risks and Regressions

- Backend integration coverage is not fully verified in this environment because the compose-backed target could not start without Docker.
- Dependency refreshes can surface latent compatibility issues later when Docker-backed integration checks are available.

## Follow-ups

- Re-run `make test-backend-integration` in an environment with Docker/Compose available.
- If lint warning noise becomes actionable later, triage the existing frontend warnings separately from the dependency refresh.

## Rollback Notes

- Revert `frontend/package.json`, `frontend/bun.lock`, `backend/go.mod`, `backend/go.sum`, and this report file to return to the prior dependency set.
