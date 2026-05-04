package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/george/pingo/internal/domain"
	"github.com/george/pingo/internal/ports/inbound"
	"github.com/george/pingo/internal/ports/outbound"
)

type ddnsService struct {
	ipFetcher   outbound.IPFetcher
	dnsProvider outbound.DNSProvider
	logger      *slog.Logger
}

// NewDDNSService creates a new DDNSService. If logger is nil, slog.Default() is used.
func NewDDNSService(ipFetcher outbound.IPFetcher, dnsProvider outbound.DNSProvider, logger *slog.Logger) inbound.DDNSService {
	if logger == nil {
		logger = slog.Default()
	}
	return &ddnsService{
		ipFetcher:   ipFetcher,
		dnsProvider: dnsProvider,
		logger:      logger,
	}
}

func (s *ddnsService) UpdateDomains(ctx context.Context, configs []domain.DomainConfig) error {
	// 1. Fetch current public IPs
	ipv4, err4 := s.ipFetcher.GetIPv4(ctx)
	if err4 != nil {
		s.logger.WarnContext(ctx, "failed to fetch IPv4", "err", err4)
	} else {
		s.logger.InfoContext(ctx, "fetched current IPv4", "ip", ipv4)
	}

	ipv6, err6 := s.ipFetcher.GetIPv6(ctx)
	if err6 != nil {
		s.logger.WarnContext(ctx, "failed to fetch IPv6", "err", err6)
	} else {
		s.logger.InfoContext(ctx, "fetched current IPv6", "ip", ipv6)
	}

	if ipv4 == "" && ipv6 == "" {
		return errors.New("failed to fetch both IPv4 and IPv6 addresses")
	}

	// 2. Process each domain configuration; aggregate per-domain failures.
	var errs []error
	for _, config := range configs {
		var currentIP string
		switch config.IPType {
		case domain.IPv4:
			currentIP = ipv4
		case domain.IPv6:
			currentIP = ipv6
		}

		if currentIP == "" {
			s.logger.InfoContext(ctx, "skipping domain because IP is unavailable",
				"domain", config.Name, "ip_type", string(config.IPType))
			continue
		}

		if err := s.processDomain(ctx, config, currentIP); err != nil {
			s.logger.ErrorContext(ctx, "failed to process domain",
				"domain", config.Name, "ip_type", string(config.IPType), "err", err)
			errs = append(errs, fmt.Errorf("%s (%s): %w", config.Name, config.IPType, err))
		}
	}

	return errors.Join(errs...)
}

func (s *ddnsService) processDomain(ctx context.Context, config domain.DomainConfig, currentIP string) error {
	recordType := config.IPType.RecordType()

	records, err := s.dnsProvider.GetRecords(ctx, config.Name, recordType)
	if err != nil {
		return fmt.Errorf("failed to get records: %w", err)
	}

	if len(records) == 0 {
		s.logger.InfoContext(ctx, "creating DNS record",
			"domain", config.Name, "type", recordType, "content", currentIP, "proxied", config.Proxied)
		return s.dnsProvider.CreateRecord(ctx, config.Name, recordType, currentIP, config.Proxied)
	}

	record := records[0]

	if len(records) > 1 {
		s.logger.WarnContext(ctx, "multiple matching records found; updating only the first",
			"domain", config.Name, "type", recordType, "id", record.ID, "count", len(records))
	}

	if record.Content == currentIP && record.Proxied == config.Proxied {
		s.logger.InfoContext(ctx, "record already up to date",
			"domain", config.Name, "type", recordType, "content", currentIP, "proxied", config.Proxied)
		return nil
	}

	s.logger.InfoContext(ctx, "updating DNS record",
		"domain", config.Name, "type", recordType, "content", currentIP, "proxied", config.Proxied, "id", record.ID)
	return s.dnsProvider.UpdateRecord(ctx, record.ID, config.Name, recordType, currentIP, config.Proxied)
}
