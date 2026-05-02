---
title: "Hexagonal architecture migration"
status: "In progress"
created: "2026-05-02"
updated: "2026-05-02"
updated_by: "george"
tags: ["architecture", "hex", "refactor"]
---

# Hexagonal architecture migration

## Current layout

```
internal/
  core/
    domain/    — entity types, no external deps
    ports/     — IPFetcher, DNSProvider (outbound), DDNSService (inbound) — not yet split
    services/  — DDNSService implementation
  adapters/
    driven/
      cloudflare/  — DNSProvider implementation
      ipfetcher/   — IPFetcher implementation
```

The `internal/core/` wrapper is a legacy nesting; the target is a flat
`internal/{domain,ports/{inbound,outbound},app,adapters}` layout.

## Migration steps

1. **Split `core/ports/` into `inbound/` and `outbound/`** — move `DDNSService`
   to `ports/inbound/`, move `IPFetcher` and `DNSProvider` to `ports/outbound/`.
   Update all imports. One PR.

2. **Create `internal/ports/` as the canonical location** — move
   `internal/core/ports/` to `internal/ports/`. Flatten the `core/` wrapper.
   Update imports throughout. One PR.

3. **Rename `core/services/` → `app/`** — move `internal/core/services/` to
   `internal/app/`. Update imports. One PR.

4. **Move `core/domain/` → `domain/`** — flatten the last `core/` remnant.
   One PR.

5. **Move `adapters/driven/` → `adapters/`** — rename to match the canonical
   `internal/adapters/<vendor>/` layout. One PR.

6. **Add function-field fakes** — add `FakeIPFetcher` and `FakeDNSProvider` to
   `internal/testdoubles/`, add fields to `ServerDeps`, remove the
   bootstrap-step comments in `deps.go`.

7. **Tighten depguard** — once the `core/` wrapper is gone, the path-based
   deny rules update to the flat paths and the adapters-no-cross-import rule
   can use the canonical glob.

## Depguard notes

Bootstrap rules active: `core-domain-no-other-internal`, `core-ports-no-impl`,
`core-services-no-adapters`, `adapters-no-cross-import`. All pass on current code.

No legacy allow entries required; existing code is already architecturally clean.
