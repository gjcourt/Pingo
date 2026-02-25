package ports

import (
	"context"

	"github.com/george/pingo/internal/core/domain"
)

// IPFetcher defines the interface for retrieving the current public IP address.
type IPFetcher interface {
	// GetIPv4 retrieves the current public IPv4 address.
	GetIPv4(ctx context.Context) (string, error)
	// GetIPv6 retrieves the current public IPv6 address.
	GetIPv6(ctx context.Context) (string, error)
}

// DNSProvider defines the interface for interacting with the DNS provider (e.g., Cloudflare).
type DNSProvider interface {
	// GetRecords retrieves all DNS records for a given domain name and record type.
	GetRecords(ctx context.Context, domainName string, recordType string) ([]domain.DNSRecord, error)
	// CreateRecord creates a new DNS record.
	CreateRecord(ctx context.Context, domainName string, recordType string, content string, proxied bool) error
	// UpdateRecord updates an existing DNS record.
	UpdateRecord(ctx context.Context, recordID string, domainName string, recordType string, content string, proxied bool) error
}

// DDNSService defines the driving port for the DDNS updater application.
type DDNSService interface {
	// UpdateDomains updates the DNS records for the configured domains.
	UpdateDomains(ctx context.Context, configs []domain.DomainConfig) error
}
