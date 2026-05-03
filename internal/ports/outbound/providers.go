// Package outbound holds the driven-port interfaces: contracts that adapters must satisfy.
package outbound

import (
	"context"

	"github.com/george/pingo/internal/domain"
)

// IPFetcher retrieves the current public IP address.
type IPFetcher interface {
	GetIPv4(ctx context.Context) (string, error)
	GetIPv6(ctx context.Context) (string, error)
}

// DNSProvider manages DNS records at a provider (e.g. Cloudflare).
type DNSProvider interface {
	GetRecords(ctx context.Context, domainName string, recordType string) ([]domain.DNSRecord, error)
	CreateRecord(ctx context.Context, domainName string, recordType string, content string, proxied bool) error
	UpdateRecord(ctx context.Context, recordID string, domainName string, recordType string, content string, proxied bool) error
}
