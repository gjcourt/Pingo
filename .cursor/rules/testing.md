# Testing Guidelines

- Write unit tests for all core logic and services.
- Use table-driven tests where applicable.
- Mock external dependencies (adapters) using interfaces.
- Aim for high test coverage, especially in the domain and service layers.
- Use the standard `testing` package.
- Test files should be named `*_test.go` and placed in the same directory as the code they test.
- Use `_test` package suffix for black-box testing (e.g., `package domain_test`).
