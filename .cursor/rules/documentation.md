---
description: Rules for maintaining project documentation
globs: ["**/*.go", "Makefile", "Dockerfile", "scripts/**", "docs/**", "README.md"]
---
# Documentation Maintenance

- Always keep the root `README.md` and the `docs/` folder up to date with the latest codebase changes.
- When adding new features, commands (e.g., Makefile targets), or environment variables, ensure they are documented in the appropriate places.
- When modifying architecture or core logic, update the corresponding architectural documentation in `docs/`.
- Ensure code examples in documentation match the current implementation.
- Before submitting a Pull Request, verify that all relevant documentation reflects the changes made in the PR.
