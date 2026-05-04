---
title: "Hex migration status"
status: "Complete"
created: "2026-05-02"
updated: "2026-05-02"
updated_by: "george"
tags: ["architecture", "hex", "tracking"]
---

# Hex migration status

## Depguard rules

| Rule | Status | Notes |
|---|---|---|
| `domain-no-other-internal` | Active ✓ | Domain is clean |
| `ports-no-impl` | Active ✓ | Ports only import domain |
| `app-no-adapters` | Active ✓ | App layer depends only on ports |
| `adapters-no-cross-import` | Active ✓ | No cross-adapter imports |

## Migration checklist

- [x] Step 1 — split `core/ports/` into `inbound/` + `outbound/`
- [x] Step 2 — promote `core/ports/` → `internal/ports/`
- [x] Step 3 — rename `core/services/` → `internal/app/`
- [x] Step 4 — move `core/domain/` → `internal/domain/`
- [x] Step 5 — rename `adapters/driven/` → `internal/adapters/`
- [x] Step 6 — add function-field fakes to `testdoubles/`
- [x] Step 7 — tighten depguard rules to flat paths
