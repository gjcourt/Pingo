# Hexagonal Architecture

This project follows the Hexagonal Architecture (Ports and Adapters) pattern.

- **Domain (`internal/core/domain`)**: Core business logic and entities. No external dependencies.
- **Ports (`internal/core/ports`)**: Interfaces defining how the application interacts with the outside world (Driven/Driving).
- **Services (`internal/core/services`)**: Application logic implementing driving ports and using driven ports.
- **Adapters (`internal/adapters`)**: Implementations of ports (e.g., Cloudflare API client, CLI entrypoint).

**Rules:**
- The `domain` package must not import any other packages from the project.
- The `ports` package can only import the `domain` package.
- The `services` package can import `domain` and `ports`.
- The `adapters` package can import `domain` and `ports`, but must not import `services`.
- The `cmd` package (entrypoint) wires everything together and can import all packages.
