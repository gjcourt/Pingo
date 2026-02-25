# Pingo Documentation

Welcome to the Pingo documentation. Pingo is a lightweight, fast, and reliable Dynamic DNS (DDNS) updater for Cloudflare, written in Go.

## Table of Contents

1. [Installation](#installation)
2. [Configuration](#configuration)
3. [Usage](#usage)
4. [Architecture](#architecture)

## Installation

### Prerequisites

- Go 1.21 or later
- A Cloudflare account with an API Token that has `Zone:DNS:Edit` permissions.

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/george/pingo.git
   cd pingo
   ```

2. Build the application:
   ```bash
   make build
   ```

   The compiled binary will be available in the `bin/` directory.

## Configuration

Pingo is configured entirely through environment variables.

| Variable | Description | Required | Default |
|---|---|---|---|
| `CLOUDFLARE_API_TOKEN` | Your Cloudflare API Token with `Zone:DNS:Edit` permissions. | Yes | None |
| `DOMAINS` | A comma-separated list of domains to update (e.g., `example.com,sub.example.com`). | Yes | None |
| `PROXIED` | Whether the DNS records should be proxied through Cloudflare (orange cloud). Set to `true` or `1` to enable. | No | `false` |

## Usage

Once configured, you can run Pingo directly:

```bash
export CLOUDFLARE_API_TOKEN="your-api-token"
export DOMAINS="example.com,sub.example.com"
export PROXIED="true"

./bin/pingo
```

Pingo will:
1. Fetch your current public IPv4 and IPv6 addresses using Cloudflare's trace endpoint (`1.1.1.1/cdn-cgi/trace`).
2. Check the existing DNS records for the configured domains.
3. Create missing records or update existing ones if the IP address has changed.

### Running as a Cron Job

Pingo is designed to be run periodically. You can set it up as a cron job to keep your DNS records up to date.

Example crontab entry (runs every 5 minutes):

```cron
*/5 * * * * CLOUDFLARE_API_TOKEN="your-api-token" DOMAINS="example.com" PROXIED="true" /path/to/pingo/bin/pingo >> /var/log/pingo.log 2>&1
```

## Architecture

Pingo is built using Hexagonal Architecture (Ports and Adapters). This design separates the core business logic from external dependencies, making the application highly testable and maintainable.

- **Domain**: Contains the core business models (`IPVersion`, `DomainConfig`, `DNSRecord`).
- **Ports**: Defines interfaces for interacting with the outside world (`IPFetcher`, `DNSProvider`, `DDNSService`).
- **Services**: Implements the application logic (`ddnsService`), orchestrating the flow between ports.
- **Adapters**: Implements the ports to interact with external systems (e.g., Cloudflare API, HTTP requests for IP fetching).
