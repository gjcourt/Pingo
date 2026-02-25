---
description: Rules for pre-push checks
globs: ["**/*"]
---
# Pre-Push Checks

- Before committing and pushing any changes, you MUST run the local CI checks to ensure nothing is broken.
- Run `make format` to ensure all code is properly formatted.
- Run `make lint` to ensure there are no linting errors.
- Run `make test` to ensure all unit tests pass.
- Do not push changes if any of these checks fail. Fix the issues first.
