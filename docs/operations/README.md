# operations/

Runbooks, smoke tests, and on-call procedures.

**Put here:**
- How to run, deploy, restart, and troubleshoot the service.
- Step-by-step recovery procedures for known failure modes.
- Smoke tests an operator can run to verify health.

**Do not put here:**
- Vendor API quirks or integration specs — `reference/`.
- Architecture explanation — `architecture/`.
- Postmortem write-ups — link from a runbook here, but the postmortem itself lives in the incident management tool.

**Naming convention:** `<yyyy-mm-dd>-<topic>.md`
Examples: `2026-05-02-running-pingo.md`, `2026-09-01-cron-misfire-recovery.md`.

**Allowed `status:` values:** `Stable`, `Superseded`.

Stale runbooks are dangerous. When a procedure changes, update the doc in the same PR.
