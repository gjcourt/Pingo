# Pingo

Pingo is a Dynamic DNS updater for Cloudflare, written in Go. Each run looks up
the host's current public IPv4/IPv6 address and reconciles a set of `A`/`AAAA`
records to match it — creating records that don't exist yet and updating ones
that have drifted.

It can optionally push the same result into one or more AdGuard Home instances
as DNS rewrites, for split-horizon setups where an internal wildcard would
otherwise shadow the public record with a stale one.

## Quickstart

```bash
make build
```

produces `bin/pingo`. Run a single reconcile:

```bash
export CLOUDFLARE_API_TOKEN="your-api-token"
export DOMAINS="example.com,sub.example.com"
export PROXIED="true"  # optional, defaults to false

./bin/pingo
```

Pingo isn't a daemon — it does one reconcile and exits. It's meant to be
re-run on a schedule (cron, systemd timer, Kubernetes CronJob).

## Configuration

Pingo is configured entirely through environment variables:

| Variable | Required | Description |
|---|---|---|
| `CLOUDFLARE_API_TOKEN` | yes | Cloudflare API token with `Zone:DNS:Edit` permission. |
| `DOMAINS` | yes | Comma-separated domains/subdomains to keep in sync. |
| `PROXIED` | no | `true`/`1` to proxy records through Cloudflare (orange cloud). Defaults to `false`. |
| `ADGUARD_URLS` | no | Comma-separated AdGuard Home base URLs. When set, `DOMAINS` are also synced there as DNS rewrites. |
| `ADGUARD_USERNAME` / `ADGUARD_PASSWORD` | no | Basic-auth credentials for the AdGuard Home control API. |

See [`docs/reference/2026-05-02-configuration.md`](docs/reference/2026-05-02-configuration.md)
for the full details, including caveats around AdGuard HA pairs and why synced
domains should stay DNS-only.

## How it works

- Public IPv4/IPv6 is discovered once per run from Cloudflare's trace endpoint
  (`1.1.1.1/cdn-cgi/trace`).
- Every `(provider, domain)` pair is reconciled concurrently; one provider
  failing doesn't block the others.
- The code follows a hexagonal (ports & adapters) layout: `internal/domain`
  holds plain types, `internal/ports` defines the `DDNSService` / `IPFetcher`
  / `DNSProvider` interfaces, `internal/app` orchestrates them, and
  `internal/adapters/{cloudflare,adguard,ipfetcher}` implement the outbound
  ports.

See [`docs/architecture/ARCHITECTURE.md`](docs/architecture/ARCHITECTURE.md)
for the component diagram and a step-by-step run-flow walkthrough.

## Development

```bash
make format   # gofmt + goimports
make lint     # golangci-lint
make test     # go test -v -race -cover ./...
make build    # -> bin/pingo
make all      # format, lint, test, build, in order
```

`internal/testdoubles` provides fakes for `IPFetcher` and `DNSProvider` used
throughout the test suite.

## Deployment

CI (`.github/workflows/image.yml`) builds a multi-arch image from the
`Dockerfile` and pushes it to `ghcr.io/gjcourt/pingo` on every push to `main`,
tagged with the UTC date, `<date>-<sha>`, and `latest`. `make image` is a
manual fallback that does the same build locally.

In the homelab, Pingo runs as a Kubernetes CronJob; the manifests live in the
`homelab` repo under `infra/controllers/pingo/` and pin a specific date tag,
so a new image push doesn't redeploy on its own.
