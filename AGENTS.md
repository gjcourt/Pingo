# AGENTS.md

> Pingo is a lightweight Dynamic DNS (DDNS) updater for Cloudflare, written in Go. — https://github.com/gjcourt/Pingo

## Commands

| Command | Use |
|---------|-----|
| `make build` | Compile to `bin/pingo` |
| `make test` | Run tests with race detector and coverage |
| `make lint` | Run golangci-lint |
| `make format` | gofmt + goimports |
| `make all` | format + lint + test + build |
| `make image` | Build and push the container image |
| `make list-images` | List published image tags |

Single test: `go test ./internal/core/services -run TestX -v`
Pre-push: `make all`

## Architecture

Hexagonal architecture (ports & adapters). Entry point: `cmd/ddns/main.go`.

- `internal/core/domain/` — entity types (`IPVersion`, `DomainConfig`, `DNSRecord`). No external deps.
- `internal/core/ports/` — interfaces (`IPFetcher`, `DNSProvider`, `DDNSService`).
- `internal/core/services/` — application orchestration (`ddnsService`).
- `internal/adapters/driven/<vendor>/` — outbound adapters (Cloudflare API client, HTTP IP fetcher).

See `docs/architecture/` for the full guide.

## Conventions

- **Core domain has no external deps** — keep `internal/core/` free of third-party libs.
- **Cloudflare SDK only in adapters** — never import `github.com/cloudflare/cloudflare-go` from `core/`.
- **New IP-detection strategies or DNS providers** implement the relevant port interface — no direct calls in `services/`.
- **Conventional Commits** for every commit (`feat:`, `fix:`, `chore:`, `refactor:`, `docs:`, `test:`, `ci:`).
- **Branch names** follow `<type>/<description>`.

## Invariants

- `internal/core/` must not import from `internal/adapters/`.
- `internal/core/` must not import any third-party packages outside stdlib.
- Cloudflare SDK types appear only in `internal/adapters/`, never in `core/`.
- The compiled binary lives at `bin/pingo`; never committed.

## What NOT to Do

- Do not add Cloudflare SDK types to `services/` or `domain/` — adapters translate, core stays pure.
- Do not skip the pre-push checks; `make all` must be green before opening a PR.
- Do not commit `bin/` artifacts or local credentials.

## Domain

A scheduled run fetches the host's public IPv4/IPv6 from Cloudflare's trace endpoint (`1.1.1.1/cdn-cgi/trace`), then reconciles each configured domain's `A`/`AAAA` records via the Cloudflare API — creating missing records or updating stale ones. Configuration is environment-variable driven; deployment is a Kubernetes CronJob.

## Cross-service dependencies

| Service | Interface | Purpose |
|---------|-----------|---------|
| Cloudflare API | `core/ports.DNSProvider` | DNS record CRUD |
| Cloudflare trace endpoint | `core/ports.IPFetcher` | Public IP discovery |

Deployed in the homelab cluster as a CronJob (`../homelab/infra/controllers/pingo/`); image-tag bumps must be coordinated with that deployment.

## Quality gate before push

1. `make format`
2. `make lint`
3. `make test`
4. `make build`

Or `make all`, which runs them in order.

## Documentation

`docs/` taxonomy: `architecture/` · `design/` · `operations/` · `plans/` · `reference/` · `research/`. See each folder's `README.md` for scope. Index: `docs/README.md`.

## Observability

Logs to stderr in slog text format at debug level. No metrics endpoint today; cluster-level CronJob status is the source of health signal. Alarms and dashboards live in the homelab observability stack — see `../homelab/docs/`.

When you learn a new convention or invariant in this repo, update this file.
