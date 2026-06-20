package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/george/pingo/internal/domain"
	"github.com/george/pingo/internal/ports/inbound"
	"github.com/george/pingo/internal/ports/outbound"
)

// NamedProvider pairs a DNS provider with a label used in logs and error
// messages, so failures across multiple sinks (e.g. Cloudflare vs AdGuard) are
// distinguishable.
type NamedProvider struct {
	Name     string
	Provider outbound.DNSProvider
}

type ddnsService struct {
	ipFetcher outbound.IPFetcher
	providers []NamedProvider
	logger    *slog.Logger
}

// NewDDNSService creates a DDNSService that reconciles against a single DNS
// provider. If logger is nil, slog.Default() is used.
func NewDDNSService(ipFetcher outbound.IPFetcher, dnsProvider outbound.DNSProvider, logger *slog.Logger) inbound.DDNSService {
	return NewDDNSServiceMulti(ipFetcher, []NamedProvider{{Name: "default", Provider: dnsProvider}}, logger)
}

// NewDDNSServiceMulti creates a DDNSService that reconciles every domain against
// each configured provider. Providers are updated independently and in
// parallel; one provider's failure does not block the others.
func NewDDNSServiceMulti(ipFetcher outbound.IPFetcher, providers []NamedProvider, logger *slog.Logger) inbound.DDNSService {
	if logger == nil {
		logger = slog.Default()
	}
	return &ddnsService{
		ipFetcher: ipFetcher,
		providers: providers,
		logger:    logger,
	}
}

func (s *ddnsService) UpdateDomains(ctx context.Context, configs []domain.DomainConfig) error {
	// 1. Fetch current public IPs once; all providers/domains share them.
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

	// 2. Reconcile every (provider, domain) pair concurrently. Each unit is
	// independent — distinct provider state and/or distinct DNS records — so
	// there is no shared mutable state on the write path beyond error
	// collection, which is mutex-guarded.
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	for _, np := range s.providers {
		for _, config := range configs {
			currentIP := s.ipForType(config.IPType, ipv4, ipv6)
			if currentIP == "" {
				s.logger.InfoContext(ctx, "skipping domain because IP is unavailable",
					"provider", np.Name, "domain", config.Name, "ip_type", string(config.IPType))
				continue
			}

			wg.Add(1)
			go func(np NamedProvider, config domain.DomainConfig, currentIP string) {
				defer wg.Done()
				if err := s.processDomain(ctx, np.Provider, config, currentIP); err != nil {
					s.logger.ErrorContext(ctx, "failed to process domain",
						"provider", np.Name, "domain", config.Name, "ip_type", string(config.IPType), "err", err)
					mu.Lock()
					errs = append(errs, fmt.Errorf("%s/%s (%s): %w", np.Name, config.Name, config.IPType, err))
					mu.Unlock()
				}
			}(np, config, currentIP)
		}
	}

	wg.Wait()
	return errors.Join(errs...)
}

func (s *ddnsService) ipForType(ipType domain.IPVersion, ipv4, ipv6 string) string {
	switch ipType {
	case domain.IPv4:
		return ipv4
	case domain.IPv6:
		return ipv6
	default:
		return ""
	}
}

func (s *ddnsService) processDomain(ctx context.Context, provider outbound.DNSProvider, config domain.DomainConfig, currentIP string) error {
	recordType := config.IPType.RecordType()

	records, err := provider.GetRecords(ctx, config.Name, recordType)
	if err != nil {
		return fmt.Errorf("failed to get records: %w", err)
	}

	if len(records) == 0 {
		s.logger.InfoContext(ctx, "creating DNS record",
			"domain", config.Name, "type", recordType, "content", currentIP, "proxied", config.Proxied)
		return provider.CreateRecord(ctx, config.Name, recordType, currentIP, config.Proxied)
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
	return provider.UpdateRecord(ctx, record.ID, config.Name, recordType, currentIP, config.Proxied)
}
