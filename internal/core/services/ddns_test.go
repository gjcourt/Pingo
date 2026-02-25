package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/george/pingo/internal/core/domain"
	"github.com/george/pingo/internal/core/services"
)

// MockIPFetcher is a mock implementation of ports.IPFetcher
type MockIPFetcher struct {
	IPv4 string
	Err4 error
	IPv6 string
	Err6 error
}

func (m *MockIPFetcher) GetIPv4(ctx context.Context) (string, error) {
	return m.IPv4, m.Err4
}

func (m *MockIPFetcher) GetIPv6(ctx context.Context) (string, error) {
	return m.IPv6, m.Err6
}

// MockDNSProvider is a mock implementation of ports.DNSProvider
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

func TestDDNSService_UpdateDomains(t *testing.T) {
	t.Run("Create missing record", func(t *testing.T) {
		ipFetcher := &MockIPFetcher{IPv4: "1.2.3.4"}
		dnsProvider := &MockDNSProvider{Records: []domain.DNSRecord{}}
		service := services.NewDDNSService(ipFetcher, dnsProvider)

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
		service := services.NewDDNSService(ipFetcher, dnsProvider)

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
		service := services.NewDDNSService(ipFetcher, dnsProvider)

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
		ipFetcher := &MockIPFetcher{Err4: errors.New("fail"), Err6: errors.New("fail")}
		dnsProvider := &MockDNSProvider{}
		service := services.NewDDNSService(ipFetcher, dnsProvider)

		configs := []domain.DomainConfig{
			{Name: "example.com", IPType: domain.IPv4, Proxied: true},
		}

		err := service.UpdateDomains(context.Background(), configs)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})
}
