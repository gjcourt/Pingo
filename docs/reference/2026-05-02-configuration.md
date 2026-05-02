---
title: Pingo configuration
status: Stable
created: 2026-05-02
updated: 2026-05-02
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

## Notes

- The API token must scope to the zones containing the listed domains. A multi-zone token works fine.
- `PROXIED` applies uniformly to all configured domains — there is no per-domain override today. If that becomes necessary, raise it as a `design/` proposal.
- Subdomains must already exist as zones or as child records under a parent zone Pingo can access.
