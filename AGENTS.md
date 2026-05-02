# Pingo Agent Guidelines

## Repository Overview

Pingo is a lightweight Dynamic DNS (DDNS) updater for Cloudflare, written in Go. It automatically detects the host's public IPv4 and IPv6 addresses and updates Cloudflare DNS `A`/`AAAA` records accordingly. Built with hexagonal architecture and zero core dependencies.

## Project Structure

```
cmd/ddns/          ← entry point
internal/          ← core domain and use cases
ddns/              ← DDNS domain types and interfaces
docs/              ← full usage and configuration docs
scripts/           ← helper scripts
bin/               ← compiled output (gitignored)
```

## Common Commands

```bash
make build         # compile to bin/pingo
make test          # run tests with race detector and coverage
make lint          # run golangci-lint
make format        # run gofmt
make all           # format + lint + test + build
```

## Architecture Guidelines

- **Core domain** has no external dependencies — keep `internal/` free of third-party libs.
- Cloudflare integration uses the official Cloudflare Go SDK in the adapter layer only.
- New IP-detection strategies or DNS providers should implement the relevant domain interfaces.

## Configuration

Requires a Cloudflare API Token with `Zone:DNS:Edit` permissions. See `docs/README.md` for full configuration options (env vars / config file).

## Deployment Notes

Pingo is deployed in the homelab cluster as a CronJob (`../homelab/infra/controllers/pingo/`). Image changes should be coordinated with the homelab deployment.
