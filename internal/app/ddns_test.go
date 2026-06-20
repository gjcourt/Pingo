package app_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/george/pingo/internal/app"
	"github.com/george/pingo/internal/domain"
)

// MockIPFetcher is a test double for outbound.IPFetcher.
type MockIPFetcher struct {
	IPv4       string
	GetIPv4Err error
	IPv6       string
	GetIPv6Err error
}

func (m *MockIPFetcher) GetIPv4(ctx context.Context) (string, error) {
	return m.IPv4, m.GetIPv4Err
}

func (m *MockIPFetcher) GetIPv6(ctx context.Context) (string, error) {
	return m.IPv6, m.GetIPv6Err
}

// MockDNSProvider is a test double for outbound.DNSProvider.
type MockDNSProvider struct {
	Records       []domain.DNSRecord
	GetErr        error
	CreateErr     error
	UpdateErr     error
	CreatedRecord *domain.DNSRecord
	UpdatedRecord *domain.DNSRecord
}

func (m *MockDNSProvider) GetRecords(ctx context.Context, domainName string, recordType string) ([]domain.DNSRecord, error) {
	return m.Records, m.GetErr
}

func (m *MockDNSProvider) CreateRecord(ctx context.Context, domainName string, recordType string, content string, proxied bool) error {
	m.CreatedRecord = &domain.DNSRecord{
		Name:    domainName,
		Type:    recordType,
		Content: content,
		Proxied: proxied,
	}
	return m.CreateErr
}

func (m *MockDNSProvider) UpdateRecord(ctx context.Context, recordID string, domainName string, recordType string, content string, proxied bool) error {
	m.UpdatedRecord = &domain.DNSRecord{
		ID:      recordID,
		Name:    domainName,
		Type:    recordType,
		Content: content,
		Proxied: proxied,
	}
	return m.UpdateErr
}

// CountingProvider is a thread-safe DNS provider double that records how many
// create/update calls it received, for exercising the concurrent fan-out under
// the race detector.
type CountingProvider struct {
	mu      sync.Mutex
	creates int
}

func (c *CountingProvider) GetRecords(_ context.Context, _ string, _ string) ([]domain.DNSRecord, error) {
	return nil, nil // always "missing" -> drives a CreateRecord
}

func (c *CountingProvider) CreateRecord(_ context.Context, _ string, _ string, _ string, _ bool) error {
	c.mu.Lock()
	c.creates++
	c.mu.Unlock()
	return nil
}

func (c *CountingProvider) UpdateRecord(_ context.Context, _ string, _ string, _ string, _ string, _ bool) error {
	return nil
}

func TestDDNSService_FansOutToAllProviders(t *testing.T) {
	ipFetcher := &MockIPFetcher{IPv4: "1.2.3.4"}
	cf := &CountingProvider{}
	ag1 := &CountingProvider{}
	ag2 := &CountingProvider{}

	service := app.NewDDNSServiceMulti(ipFetcher, []app.NamedProvider{
		{Name: "cloudflare", Provider: cf},
		{Name: "adguard-a", Provider: ag1},
		{Name: "adguard-b", Provider: ag2},
	}, nil)

	configs := []domain.DomainConfig{
		{Name: "vpn.example.com", IPType: domain.IPv4},
		{Name: "home.example.com", IPType: domain.IPv4},
	}

	if err := service.UpdateDomains(context.Background(), configs); err != nil {
		t.Fatalf("UpdateDomains() error = %v", err)
	}

	// Every provider should have been reconciled for every domain.
	for name, p := range map[string]*CountingProvider{"cloudflare": cf, "adguard-a": ag1, "adguard-b": ag2} {
		if p.creates != len(configs) {
			t.Errorf("provider %s: got %d creates, want %d", name, p.creates, len(configs))
		}
	}
}

func TestDDNSService_UpdateDomains(t *testing.T) {
	t.Run("Create missing record", func(t *testing.T) {
		ipFetcher := &MockIPFetcher{IPv4: "1.2.3.4"}
		dnsProvider := &MockDNSProvider{Records: []domain.DNSRecord{}}
		service := app.NewDDNSService(ipFetcher, dnsProvider, nil)

		configs := []domain.DomainConfig{
			{Name: "example.com", IPType: domain.IPv4, Proxied: true},
		}

		err := service.UpdateDomains(context.Background(), configs)
		if err != nil {
			t.Fatalf("UpdateDomains() error = %v", err)
		}

		if dnsProvider.CreatedRecord == nil {
			t.Fatal("Expected record to be created")
		}
		if dnsProvider.CreatedRecord.Content != "1.2.3.4" {
			t.Errorf("Created record content = %v, want %v", dnsProvider.CreatedRecord.Content, "1.2.3.4")
		}
	})

	t.Run("Update existing record", func(t *testing.T) {
		ipFetcher := &MockIPFetcher{IPv4: "1.2.3.4"}
		dnsProvider := &MockDNSProvider{
			Records: []domain.DNSRecord{
				{ID: "123", Name: "example.com", Type: "A", Content: "old.ip", Proxied: true},
			},
		}
		service := app.NewDDNSService(ipFetcher, dnsProvider, nil)

		configs := []domain.DomainConfig{
			{Name: "example.com", IPType: domain.IPv4, Proxied: true},
		}

		err := service.UpdateDomains(context.Background(), configs)
		if err != nil {
			t.Fatalf("UpdateDomains() error = %v", err)
		}

		if dnsProvider.UpdatedRecord == nil {
			t.Fatal("Expected record to be updated")
		}
		if dnsProvider.UpdatedRecord.Content != "1.2.3.4" {
			t.Errorf("Updated record content = %v, want %v", dnsProvider.UpdatedRecord.Content, "1.2.3.4")
		}
	})

	t.Run("No update needed", func(t *testing.T) {
		ipFetcher := &MockIPFetcher{IPv4: "1.2.3.4"}
		dnsProvider := &MockDNSProvider{
			Records: []domain.DNSRecord{
				{ID: "123", Name: "example.com", Type: "A", Content: "1.2.3.4", Proxied: true},
			},
		}
		service := app.NewDDNSService(ipFetcher, dnsProvider, nil)

		configs := []domain.DomainConfig{
			{Name: "example.com", IPType: domain.IPv4, Proxied: true},
		}

		err := service.UpdateDomains(context.Background(), configs)
		if err != nil {
			t.Fatalf("UpdateDomains() error = %v", err)
		}

		if dnsProvider.UpdatedRecord != nil {
			t.Fatal("Expected record NOT to be updated")
		}
	})

	t.Run("Fails if both IPs fail", func(t *testing.T) {
		ipFetcher := &MockIPFetcher{GetIPv4Err: errors.New("fail"), GetIPv6Err: errors.New("fail")}
		dnsProvider := &MockDNSProvider{}
		service := app.NewDDNSService(ipFetcher, dnsProvider, nil)

		configs := []domain.DomainConfig{
			{Name: "example.com", IPType: domain.IPv4, Proxied: true},
		}

		err := service.UpdateDomains(context.Background(), configs)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})
}
