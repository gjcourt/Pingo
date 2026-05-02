# Pingo Documentation

Welcome to the Pingo documentation. Pingo is a lightweight Dynamic DNS (DDNS) updater for Cloudflare, written in Go.

This folder is organized into a fixed six-folder taxonomy. Each folder's `README.md` describes what belongs there.

## Folders

- [`architecture/`](architecture/README.md) — how the system is built today.
- [`design/`](design/README.md) — proposals, RFCs, in-flight or recently shipped designs.
- [`operations/`](operations/README.md) — runbooks, smoke tests, deployment and on-call procedures.
- [`plans/`](plans/README.md) — phased migrations, rollout sequencing.
- [`reference/`](reference/README.md) — configuration, vendor specs, lookup material.
- [`research/`](research/README.md) — spikes, investigations, vendor evaluations.

## Quick links

- **Want to run Pingo?** → [`operations/2026-05-02-running-pingo.md`](operations/2026-05-02-running-pingo.md).
- **Configuration env vars?** → [`reference/2026-05-02-configuration.md`](reference/2026-05-02-configuration.md).
- **Architecture overview?** → [`architecture/2026-05-02-hexagonal-overview.md`](architecture/2026-05-02-hexagonal-overview.md).

## Conventions

- All non-index docs use frontmatter (`title`, `status`, `created`, `updated`, `updated_by`, `tags`). See [`reference/2026-05-02-configuration.md`](reference/2026-05-02-configuration.md) for an example.
- Filenames carry topic and creation date (`<yyyy-mm-dd>-<topic>.md`); state lives in `status:` frontmatter, never in the filename.
- See `AGENTS.md` for the full convention.
