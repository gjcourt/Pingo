# Pingo

Pingo is a lightweight, fast, and reliable Dynamic DNS (DDNS) updater for Cloudflare, written in Go. It automatically detects your public IPv4 and IPv6 addresses and updates your Cloudflare DNS records accordingly.

## Features

- **IPv4 and IPv6 Support**: Automatically updates both `A` and `AAAA` records.
- **Cloudflare Integration**: Uses the official Cloudflare Go SDK for robust API interactions.
- **Proxy Configuration**: Easily configure whether your DNS records should be proxied (orange cloud) or DNS-only.
- **Hexagonal Architecture**: Built with clean architecture principles for maintainability and testability.
- **Zero Dependencies**: The core logic has no external dependencies, making it easy to test and extend.

## Getting Started

See the [Documentation](docs/README.md) for detailed instructions on how to install, configure, and run Pingo.

## Quick Start

1. Get a Cloudflare API Token with `Zone:DNS:Edit` permissions.
2. Run the application:

```bash
export CF_API_TOKEN="your-api-token"
export DOMAINS="example.com,sub.example.com"
export PROXIED="true" # Optional, defaults to false

./bin/pingo
```

## Development

Pingo uses a standard Go toolchain.

```bash
# Build the application
make build

# Run tests
make test

# Run linters
make lint

# Format code
make format
```

## License

MIT License
