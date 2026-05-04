---
title: Pingo hexagonal architecture overview
status: Stable
created: 2026-05-02
updated: 2026-05-03
updated_by: gjcourt
tags: [architecture, hexagonal]
---

# Pingo hexagonal architecture overview

Pingo is built using hexagonal architecture (ports and adapters). This separates the core business logic from external dependencies, making the application highly testable and replaceable at the boundary.

## Layers

- **Domain** (`internal/domain/`) — entity types and value objects: `IPVersion`, `DomainConfig`, `DNSRecord`. No external dependencies.
- **Inbound ports** (`internal/ports/inbound/`) — driving-port interfaces: `DDNSService`.
- **Outbound ports** (`internal/ports/outbound/`) — driven-port interfaces: `IPFetcher`, `DNSProvider`.
- **Application** (`internal/app/`) — application logic that orchestrates the flow between ports (`ddnsService`).
- **Adapters** (`internal/adapters/<vendor>/`) — concrete implementations of the outbound ports that talk to external systems (Cloudflare API client at `internal/adapters/cloudflare/`, HTTP-based public-IP fetcher at `internal/adapters/ipfetcher/`).
- **Entry point** (`cmd/ddns/main.go`) — wires the adapters into the service and runs a single reconcile pass.

## Dependency rule

Imports flow inward only:

```
adapters/  →  ports/  →  domain/
app/       →  ports/, domain/
```

Nothing in `domain/`, `ports/`, or `app/` may import from `adapters/`. Third-party packages (e.g. the Cloudflare SDK) appear only in `adapters/`.

## Adding a new DNS provider or IP-detection strategy

1. Define (or reuse) the relevant interface in `internal/ports/outbound/`.
2. Place the implementation under `internal/adapters/<name>/`.
3. Wire it in `cmd/ddns/main.go`.
4. Add tests against the port interface, not the adapter directly.
