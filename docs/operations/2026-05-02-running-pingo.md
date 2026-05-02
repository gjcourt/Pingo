---
title: Running Pingo
status: Stable
created: 2026-05-02
updated: 2026-05-02
updated_by: gjcourt
tags: [operations, deployment, cron]
---

# Running Pingo

## Prerequisites

- Go 1.21 or later (Go toolchain pinned in `go.mod`).
- A Cloudflare account with an API Token that has `Zone:DNS:Edit` permissions.

## Build

```bash
git clone https://github.com/gjcourt/Pingo.git
cd Pingo
make build
```

The compiled binary lands at `bin/pingo`.

## Run locally

Pingo is configured entirely through environment variables — see `../reference/2026-05-02-configuration.md` for the full list.

```bash
export CLOUDFLARE_API_TOKEN="your-api-token"
export DOMAINS="example.com,sub.example.com"
export PROXIED="true"

./bin/pingo
```

A single run:

1. Fetches the host's public IPv4 and IPv6 addresses from Cloudflare's trace endpoint (`1.1.1.1/cdn-cgi/trace`).
2. Reads the existing DNS records for each configured domain.
3. Creates missing records or updates existing ones if the IP has changed.

## Run as a cron job

Pingo is designed for periodic execution. Example crontab entry (every 5 minutes):

```cron
*/5 * * * * CLOUDFLARE_API_TOKEN="your-api-token" DOMAINS="example.com" PROXIED="true" /path/to/pingo/bin/pingo >> /var/log/pingo.log 2>&1
```

## Run in the homelab cluster

Pingo is deployed as a Kubernetes CronJob. The deployment manifests live in the `homelab` repo at `infra/controllers/pingo/`. Image-tag bumps must be coordinated with that deployment — pushing a new image alone does not redeploy.
