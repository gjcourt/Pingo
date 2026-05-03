package app

import (
	"context"
	"fmt"
	"log"

	"github.com/george/pingo/internal/domain"
	"github.com/george/pingo/internal/ports/inbound"
	"github.com/george/pingo/internal/ports/outbound"
)

type ddnsService struct {
	ipFetcher   outbound.IPFetcher
	dnsProvider outbound.DNSProvider
}

// NewDDNSService creates a new DDNSService.
func NewDDNSService(ipFetcher outbound.IPFetcher, dnsProvider outbound.DNSProvider) inbound.DDNSService {
	return &ddnsService{
		ipFetcher:   ipFetcher,
		dnsProvider: dnsProvider,
	}
}

func (s *ddnsService) UpdateDomains(ctx context.Context, configs []domain.DomainConfig) error {
	// 1. Fetch current public IPs
	ipv4, err4 := s.ipFetcher.GetIPv4(ctx)
	if err4 != nil {
		log.Printf("Warning: Failed to fetch IPv4: %v", err4)
	} else {
		log.Printf("Current IPv4: %s", ipv4)
	}

	ipv6, err6 := s.ipFetcher.GetIPv6(ctx)
	if err6 != nil {
		log.Printf("Warning: Failed to fetch IPv6: %v", err6)
	} else {
		log.Printf("Current IPv6: %s", ipv6)
	}

	if ipv4 == "" && ipv6 == "" {
		return fmt.Errorf("failed to fetch both IPv4 and IPv6 addresses")
	}

	// 2. Process each domain configuration
	for _, config := range configs {
		var currentIP string
		switch config.IPType {
		case domain.IPv4:
			currentIP = ipv4
		case domain.IPv6:
			currentIP = ipv6
		}

		if currentIP == "" {
			log.Printf("Skipping %s (%s) because IP address is not available", config.Name, config.IPType)
			continue
		}

		err := s.processDomain(ctx, config, currentIP)
		if err != nil {
			log.Printf("Error processing domain %s: %v", config.Name, err)
		}
	}

	return nil
}

func (s *ddnsService) processDomain(ctx context.Context, config domain.DomainConfig, currentIP string) error {
	recordType := config.IPType.RecordType()

	records, err := s.dnsProvider.GetRecords(ctx, config.Name, recordType)
	if err != nil {
		return fmt.Errorf("failed to get records: %w", err)
	}

	if len(records) == 0 {
		log.Printf("Creating %s record for %s -> %s (Proxied: %t)", recordType, config.Name, currentIP, config.Proxied)
		return s.dnsProvider.CreateRecord(ctx, config.Name, recordType, currentIP, config.Proxied)
	}

	record := records[0]

	if len(records) > 1 {
		log.Printf("Warning: Found multiple %s records for %s. Updating the first one (%s).", recordType, config.Name, record.ID)
	}

	if record.Content == currentIP && record.Proxied == config.Proxied {
		log.Printf("No update needed for %s (%s). Current IP: %s, Proxied: %t", config.Name, recordType, currentIP, config.Proxied)
		return nil
	}

	log.Printf("Updating %s record for %s -> %s (Proxied: %t)", recordType, config.Name, currentIP, config.Proxied)
	return s.dnsProvider.UpdateRecord(ctx, record.ID, config.Name, recordType, currentIP, config.Proxied)
}
