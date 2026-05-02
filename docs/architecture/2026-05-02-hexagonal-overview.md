---
title: Pingo hexagonal architecture overview
status: Stable
created: 2026-05-02
updated: 2026-05-02
updated_by: gjcourt
tags: [architecture, hexagonal]
---

# Pingo hexagonal architecture overview

Pingo is built using hexagonal architecture (ports and adapters). This separates the core business logic from external dependencies, making the application highly testable and replaceable at the boundary.

## Layers

- **Domain** (`internal/core/domain/`) — entity types and value objects: `IPVersion`, `DomainConfig`, `DNSRecord`. No external dependencies.
- **Ports** (`internal/core/ports/`) — interfaces describing the boundaries: `IPFetcher`, `DNSProvider`, `DDNSService`.
- **Services** (`internal/core/services/`) — application logic that orchestrates the flow between ports (`ddnsService`).
- **Adapters** (`internal/adapters/driven/<vendor>/`) — concrete implementations of the outbound ports that talk to external systems (e.g., Cloudflare API client at `driven/cloudflare/`, HTTP-based public-IP fetcher at `driven/ipfetcher/`). The `driven/` subfolder marks these as outbound (driven) adapters in the hexagonal vocabulary.
- **Entry point** (`cmd/ddns/main.go`) — wires the adapters into the service and runs a single reconcile pass.

## Dependency rule

Imports flow inward only:

```
adapters/  →  core/ports/  →  core/domain/
core/services/  →  core/ports/, core/domain/
```

Nothing in `core/` may import `adapters/` or third-party packages outside the standard library.

## Adding a new DNS provider or IP-detection strategy

1. Define (or reuse) the relevant interface in `core/ports/`.
2. Place the implementation under `internal/adapters/driven/<name>/`.
3. Wire it in `cmd/ddns/main.go`.
4. Add tests against the port interface, not the adapter directly.
