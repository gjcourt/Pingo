# Go Best Practices

- Follow standard Go idioms and conventions (Effective Go).
- Use `gofmt` and `goimports` for formatting.
- Keep functions small and focused on a single responsibility.
- Use meaningful and concise variable names.
- Handle errors explicitly; do not ignore them.
- Avoid global state and package-level variables where possible.
- Use `context.Context` for cancellation and timeouts in long-running or blocking operations.
- Use goroutines and channels for concurrent operations. Avoid shared memory; communicate by sharing memory.
