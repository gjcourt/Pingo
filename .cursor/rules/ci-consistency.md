---
description: Rules for maintaining CI/CD consistency
globs: ["go.mod", ".github/workflows/*.yml", "Dockerfile", ".golangci.yml"]
---
# CI/CD Consistency

- **Go Version Sync**: Always ensure that the Go version is identical across `go.mod`, `.github/workflows/ci.yml`, and `Dockerfile`. If you update the Go version in one file, you MUST update it in the others.
- **Linter Compatibility**: Ensure that `.golangci.yml` is compatible with the `golangci-lint` version used in GitHub Actions (currently v1.x). Do not use configuration fields (like `version: 2`) that are unsupported by the CI runner.
- **Local vs CI**: If a CI check fails on GitHub Actions but passes locally, investigate version mismatches between the local environment and the CI environment.
