---
title: "Hex migration status"
status: "In progress"
created: "2026-05-02"
updated: "2026-05-02"
updated_by: "george"
tags: ["architecture", "hex", "tracking"]
---

# Hex migration status

## Depguard rules

| Rule | Status | Notes |
|---|---|---|
| `core-domain-no-other-internal` | Active ✓ | Domain is clean |
| `core-ports-no-impl` | Active ✓ | Ports only import domain |
| `core-services-no-adapters` | Active ✓ | Services only import domain + ports |
| `adapters-no-cross-import` | Active ✓ | No cross-adapter imports |

## Migration checklist

- [ ] Step 1 — split `core/ports/` into `inbound/` + `outbound/`
- [ ] Step 2 — promote `core/ports/` → `internal/ports/`
- [ ] Step 3 — rename `core/services/` → `internal/app/`
- [ ] Step 4 — move `core/domain/` → `internal/domain/`
- [ ] Step 5 — rename `adapters/driven/` → `internal/adapters/`
- [ ] Step 6 — add function-field fakes to `testdoubles/`
- [ ] Step 7 — tighten depguard rules to flat paths
