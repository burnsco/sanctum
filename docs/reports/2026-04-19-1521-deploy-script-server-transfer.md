# Deploy Script Server Transfer Report

## Metadata

- Date: `2026-04-19`
- Branch: `main`
- Author/Agent: `Codex`
- Scope: Update `deploy.sh` to use a server-safe direct image transfer path for `mordor`

## Structured Signals

```json
{
  "Report-Version": "1.0",
  "Domains": ["infra"],
  "Lessons": [
    {
      "title": "Registry push permissions should not block deployment",
      "severity": "MEDIUM",
      "anti_pattern": "Using a deploy script that depends on pushing to GHCR can block production updates when the active token lacks package write access.",
      "detection": "rg -n \"docker push|ghcr.io/burnsco/sanctum\" deploy.sh Makefile",
      "prevention": "Prefer a direct image transfer or a registry-independent fallback for the server deployment path."
    }
  ]
}
```

## Summary

- Replaced the GHCR push-based deploy step with a direct image transfer to `mordor`.
- Kept the local build step intact so the script still builds the frontend and backend images before deployment.
- Added a server-side safety check for `alpine:3.21`, which the production compose stack requires for the `uploads-init` service.

## Changes Made

- `deploy.sh`
  - Added `set -euo pipefail` and env-var overrides for host, port, user, and paths.
  - Added a temporary archive workflow with `docker save` and `scp`.
  - Replaced `docker push` and remote `docker compose pull` with `docker load` plus `docker compose up -d --no-build --pull never`.
  - Added a remote `alpine:3.21` pull fallback if the server does not already have that base image.

## Validation

- Commands run:
  - `docker version`
  - `make lint`
  - `make test-backend-integration`
  - `./deploy.sh`
  - `ssh -p 2222 cburns@192.168.2.124 'cd /home/cburns/docker/sanctum && docker compose ps'`
  - `bash -n deploy.sh`
- Test results:
  - Backend lint passed once Docker was available.
  - Backend integration tests passed.
  - Direct deploy path succeeded on `mordor` after the script was updated.
- Manual verification:
  - Server compose reported `sanctum-backend` and `sanctum-frontend` running with the refreshed `latest` images.

## Risks and Regressions

- The script now depends on the ability to use `scp` and `ssh` to the server.
- Large image archives temporarily consume local disk space during transfer.
- If the server compose stack changes away from `latest` image tags, the direct load path will need to be updated in tandem.

## Follow-ups

- Consider documenting the direct-transfer deployment flow in the server runbook if this becomes the standard path.
- Consider removing or repurposing the registry-push workflow if it is no longer used operationally.

## Rollback Notes

- Restore the previous `deploy.sh` if you need to revert to the GHCR push workflow.
- If the remote stack is already updated, redeploying the prior commit or previous image archive will return the server to the known-good state.
