# Task Report: Containerized Seeding Workflow

## Metadata

- Date: `2026-03-12`
- Branch: `master`
- Author/Agent: `Codex`
- Scope: `Add containerized realistic seeding paths for local/dev and deployed production-image workflows`

## Structured Signals

```json
{
  "Report-Version": "1.0",
  "Domains": ["backend", "infra", "docs"],
  "Lessons": [
    {
      "title": "Operational seed flows need a dedicated production binary",
      "severity": "HIGH",
      "anti_pattern": "Relying on `go run` for seeding in a production image that only contains the app binary",
      "detection": "rg -n \"cmd/seed|ENTRYPOINT|target: production|seed-image\" Dockerfile Makefile docs/development/seeding.md",
      "prevention": "Build and copy a dedicated `/seed` binary into the production image for manual operational workflows."
    },
    {
      "title": "One-off compose tasks should build the intended image variant explicitly",
      "severity": "MEDIUM",
      "anti_pattern": "Dev and production compose variants sharing the same image name, causing ad-hoc commands to use the wrong runtime contents",
      "detection": "rg -n \"target: dev|target: production|seed-docker|seed-image\" compose.yml compose.override.yml Makefile",
      "prevention": "Have each one-off Make target build the exact compose variant it intends to run before invoking the command."
    }
  ]
}
```

## Summary

- Requested a way to test realistic seeding locally first, then run an equivalent seeding flow from the server against a deployed environment.
- Delivered containerized seed targets for both the dev workflow and the production-image workflow, plus docs and local validation.

## Changes Made

- Updated `Dockerfile` to build and copy a dedicated `/seed` binary into the production image alongside `/main`.
- Added Make targets in `Makefile` for:
  - `seed-docker`
  - `seed-realistic-docker`
  - `seed-realistic-docker-append`
  - `seed-image`
  - `seed-realistic-image`
  - `seed-realistic-image-append`
- Made the dev seed target explicitly build the dev image and invoke `/usr/local/go/bin/go` to avoid PATH drift in `docker compose run --entrypoint sh`.
- Made the production-image seed target explicitly build the production image before invoking `/seed`.
- Updated `docs/development/seeding.md` with local/dev commands, production-image commands, and `-clean=false` append guidance.

## Validation

- Commands run:
  - `APP_ENV=development ENVIRONMENT=prod make build-backend`
  - `APP_ENV=development make seed-realistic-image`
  - `make seed-realistic-docker-append`
  - `APP_ENV=development make seed-realistic-image`
  - `./scripts/compose.sh -f compose.yml exec -T postgres psql -U ${POSTGRES_USER:-sanctum_user} -d ${POSTGRES_DB:-sanctum_test} -c "SELECT (SELECT COUNT(*) FROM users) AS users, (SELECT COUNT(*) FROM posts) AS posts, (SELECT COUNT(*) FROM comments) AS comments, (SELECT COUNT(*) FROM likes) AS likes;"`
- Test results:
  - Production-image seed path completed successfully and produced `20` users, `134` posts, `401` comments.
  - Dev append seed path completed successfully and temporarily increased counts to `40` users, `268` posts, `802` comments before the final destructive reseed.
  - Final production-image reseed completed successfully and returned the DB to `20` users, `134` posts, `401` comments.
- Manual verification:
  - Confirmed the realistic preset logs every built-in sanctum and seeds the expected per-sanctum content distribution.
  - Confirmed the production-image target still works after the dev target rebuilds the backend image to the dev variant.

## Risks and Regressions

- Known risks:
  - `seed-realistic-image` is destructive by default because the seed CLI still defaults `-clean=true`.
  - Append mode still attempts to create seed users with fixed usernames like `cburns` and `test`, which logs duplicate-key errors before continuing.
- Potential regressions:
  - Image build time is now part of the containerized seed targets.
  - Operators could mistake `seed-image` for a safe append command if they skip the docs.
- Mitigations:
  - Added explicit `*-append` targets for non-destructive runs.
  - Documented the destructive default and the production-image workflow in the seeding guide.

## Follow-ups

- Remaining work:
  - Consider a dedicated backfill preset that reuses existing users instead of creating new seed users in append mode.
  - Consider a post-seed verification target that prints per-sanctum post counts automatically.
- Recommended next steps:
  - If you want prod to be fully demo-populated from scratch, run `make seed-realistic-image` on the server after deploying this change.
  - If you want to preserve any existing rows, use `make seed-realistic-image-append` instead.

## Rollback Notes

- Revert `Dockerfile`, `Makefile`, `docs/development/seeding.md`, and this report file.
- If a seeded environment needs to be reset after rollout, re-run the destructive seed target or restore the database from backup according to your deployment process.
