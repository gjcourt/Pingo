<!-- readme-type: service -->
# Pingo

Keeps Cloudflare (and optional AdGuard Home) DNS records in sync with your host's public IP

Cloudflare-hosted domains go stale when a host's public IP changes and nothing tells Cloudflare. Pingo runs on a schedule, discovers the host's current public IPv4/IPv6 address, and reconciles each configured domain's `A`/`AAAA` records to match it — creating records that don't exist yet and updating ones that have drifted. It can optionally push the same result into one or more AdGuard Home instances as DNS rewrites, for split-horizon setups where an internal wildcard would otherwise shadow the public record with a stale one.

**Status:** in daily use on the homelab, deployed as a Kubernetes CronJob since 2026-02.

## Quick start

Needs: Go 1.24 or later and a Cloudflare API token with `Zone:DNS:Edit` permission.

```bash
git clone https://github.com/gjcourt/Pingo.git && cd Pingo
make build
```

```bash
export CLOUDFLARE_API_TOKEN="your-api-token"
export DOMAINS="example.com,sub.example.com"
./bin/pingo
```

Pingo isn't a daemon — it does one reconcile and exits.

## Usage

Schedule it to run on an interval; for example, every five minutes via cron:

```cron
*/5 * * * * CLOUDFLARE_API_TOKEN="your-api-token" DOMAINS="example.com" /path/to/pingo/bin/pingo >> /var/log/pingo.log 2>&1
```

See [running Pingo](docs/operations/2026-05-02-running-pingo.md) for systemd timer and Kubernetes CronJob examples.

## Configuration

Pingo is configured entirely through environment variables:

| Variable | Required | Description |
|---|---|---|
| `CLOUDFLARE_API_TOKEN` | yes | Cloudflare API token with `Zone:DNS:Edit` permission. |
| `DOMAINS` | yes | Comma-separated domains/subdomains to keep in sync. |
| `PROXIED` | no | `true`/`1` to proxy records through Cloudflare (orange cloud). Defaults to `false`. |
| `ADGUARD_URLS` | no | Comma-separated AdGuard Home base URLs. When set, `DOMAINS` are also synced there as DNS rewrites. |
| `ADGUARD_USERNAME` / `ADGUARD_PASSWORD` | no | Basic-auth credentials for the AdGuard Home control API. |

See the [configuration reference](docs/reference/2026-05-02-configuration.md) for the full details, including caveats around AdGuard HA pairs and why synced domains should stay DNS-only.

## How it works

Each run discovers the host's public IPv4/IPv6 from Cloudflare's trace endpoint (`1.1.1.1/cdn-cgi/trace`), then reconciles every configured `(provider, domain)` pair concurrently, so one provider failing doesn't block the others. The code follows a hexagonal (ports & adapters) layout — see [ARCHITECTURE.md](docs/architecture/ARCHITECTURE.md) for the component diagram and a full run-flow walkthrough.

## Development

```bash
make format   # gofmt + goimports
make lint     # golangci-lint
make test     # go test -v -race -cover ./...
make build    # -> bin/pingo
make all      # format, lint, test, build, in order
```

CI also runs `go-arch-lint check` to enforce the hexagonal dependency rule. Conventions for contributors and agents: [AGENTS.md](AGENTS.md).

## Deployment

CI (`.github/workflows/image.yml`) builds a multi-arch image and pushes it to `ghcr.io/gjcourt/pingo` on every push to `main`, tagged with the UTC date, `<date>-<sha>`, and `latest`. `make image` is a manual fallback that does the same build locally.

In the homelab, Pingo runs as a Kubernetes CronJob — see the
[runbook](https://github.com/gjcourt/homelab/blob/master/docs/reference/pingo.md) in `gjcourt/homelab`. Image-tag bumps must be coordinated with that deployment; a new image push doesn't redeploy on its own.

## License

No licence file yet.
