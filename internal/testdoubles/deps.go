package testdoubles

import (
	"context"

	"github.com/george/pingo/internal/domain"
	"github.com/george/pingo/internal/ports/outbound"
)

// FakeIPFetcher is a controllable outbound.IPFetcher for unit tests.
type FakeIPFetcher struct {
	IPv4       string
	GetIPv4Err error
	IPv6       string
	GetIPv6Err error
}

func (f *FakeIPFetcher) GetIPv4(_ context.Context) (string, error) {
	return f.IPv4, f.GetIPv4Err
}

func (f *FakeIPFetcher) GetIPv6(_ context.Context) (string, error) {
	return f.IPv6, f.GetIPv6Err
}

var _ outbound.IPFetcher = (*FakeIPFetcher)(nil)

// FakeDNSProvider is a controllable outbound.DNSProvider for unit tests.
type FakeDNSProvider struct {
	Records       []domain.DNSRecord
	GetErr        error
	CreateErr     error
	UpdateErr     error
	CreatedRecord *domain.DNSRecord
	UpdatedRecord *domain.DNSRecord
}

func (f *FakeDNSProvider) GetRecords(_ context.Context, _ string, _ string) ([]domain.DNSRecord, error) {
	return f.Records, f.GetErr
}

func (f *FakeDNSProvider) CreateRecord(_ context.Context, domainName, recordType, content string, proxied bool) error {
	f.CreatedRecord = &domain.DNSRecord{Name: domainName, Type: recordType, Content: content, Proxied: proxied}
	return f.CreateErr
}

func (f *FakeDNSProvider) UpdateRecord(_ context.Context, recordID, domainName, recordType, content string, proxied bool) error {
	f.UpdatedRecord = &domain.DNSRecord{ID: recordID, Name: domainName, Type: recordType, Content: content, Proxied: proxied}
	return f.UpdateErr
}

var _ outbound.DNSProvider = (*FakeDNSProvider)(nil)

// ServerDeps aggregates all outbound-port fakes for unit tests.
type ServerDeps struct {
	IPFetcher   *FakeIPFetcher
	DNSProvider *FakeDNSProvider
}

// NewServerDeps returns a ServerDeps with all fakes initialised to safe zero-value defaults.
func NewServerDeps() *ServerDeps {
	return &ServerDeps{
		IPFetcher:   &FakeIPFetcher{},
		DNSProvider: &FakeDNSProvider{},
	}
}
