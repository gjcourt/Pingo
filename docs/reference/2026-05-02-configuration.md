---
title: Pingo configuration
status: Stable
created: 2026-05-02
updated: 2026-06-19
updated_by: gjcourt
tags: [configuration, environment]
---

# Pingo configuration

Pingo is configured entirely through environment variables.

| Variable | Description | Required | Default |
|---|---|---|---|
| `CLOUDFLARE_API_TOKEN` | Cloudflare API Token with `Zone:DNS:Edit` permissions. | Yes | — |
| `DOMAINS` | Comma-separated list of domains to update (e.g., `example.com,sub.example.com`). | Yes | — |
| `PROXIED` | Whether DNS records should be proxied through Cloudflare (orange cloud). Set `true` or `1` to enable. | No | `false` |
| `ADGUARD_URLS` | Comma-separated list of AdGuard Home base URLs (e.g., `http://10.0.0.2,http://10.0.0.3`). When set, Pingo also syncs each domain as a DNS **rewrite** on every listed instance. Leave unset to disable. | No | — |
| `ADGUARD_USERNAME` | HTTP basic-auth username for the AdGuard Home control API. | No | — |
| `ADGUARD_PASSWORD` | HTTP basic-auth password for the AdGuard Home control API. | No | — |

## Notes

- The API token must scope to the zones containing the listed domains. A multi-zone token works fine.
- `PROXIED` applies uniformly to all configured domains — there is no per-domain override today. If that becomes necessary, raise it as a `design/` proposal.
- Subdomains must already exist as zones or as child records under a parent zone Pingo can access.

## AdGuard Home rewrite sync

When `ADGUARD_URLS` is set, Pingo updates the same `DOMAINS` as DNS rewrites on each AdGuard instance, in addition to Cloudflare. All providers (Cloudflare + each AdGuard) are reconciled **concurrently**.

**Why:** split-horizon DNS. If an internal wildcard (e.g. `*.example.com → <reverse-proxy>`) catches a host that must instead resolve to the live public IP for LAN clients — typically a WireGuard/VPN endpoint reached via NAT hairpin — a more-specific AdGuard rewrite overrides the wildcard. Pingo keeps that rewrite pointed at the current public IP automatically.

- List **both** instances of an HA pair in `ADGUARD_URLS` if they don't replicate rewrites between themselves; updates are idempotent.
- Rewrites have no proxy concept, so synced domains should be **DNS-only** (`PROXIED=false`) — which a VPN endpoint requires regardless (the handshake needs the real IP, not a proxied one).
- A rewrite update is a delete-then-add (AdGuard has no update verb); Pingo carries the old answer as the record identity to delete on.
