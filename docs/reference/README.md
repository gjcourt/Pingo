# reference/

Information you look things up in — vendor APIs, domain knowledge, integration specs, configuration tables.

**Put here:**
- Configuration env-var tables.
- Vendor API gotchas, request/response shapes, rate-limit tables.
- Long-shelf-life lookup material that an engineer or agent re-reads each time.

**Do not put here:**
- Runbook steps — `operations/`.
- Architecture overview — `architecture/`.
- Spike output — `research/`.

**Naming convention:** `<yyyy-mm-dd>-<topic>.md`
Examples: `2026-05-02-configuration.md`, `2026-08-01-cloudflare-api-quirks.md`.

**Allowed `status:` values:** `Stable`, `Superseded`.

Date prefix is bumped when the doc is materially revised.
