package cloudflare

import (
	"context"
	"fmt"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/george/pingo/internal/core/domain"
	"github.com/george/pingo/internal/core/ports"
)

type adapter struct {
	client *cloudflare.Client
}

// NewAdapter creates a new Cloudflare DNS provider adapter.
func NewAdapter(apiToken string) (ports.DNSProvider, error) {
	client := cloudflare.NewClient(option.WithAPIToken(apiToken))
	return &adapter{client: client}, nil
}

// getZoneID finds the Zone ID for a given domain name.
func (a *adapter) getZoneID(ctx context.Context, domainName string) (string, error) {
	parts := strings.Split(domainName, ".")
	for i := 0; i < len(parts)-1; i++ {
		zoneName := strings.Join(parts[i:], ".")
		page, err := a.client.Zones.List(ctx, zones.ZoneListParams{
			Name: cloudflare.F(zoneName),
		})
		if err == nil && len(page.Result) > 0 {
			return page.Result[0].ID, nil
		}
	}
	return "", fmt.Errorf("could not find zone for domain: %s", domainName)
}

func (a *adapter) GetRecords(ctx context.Context, domainName string, recordType string) ([]domain.DNSRecord, error) {
	zoneID, err := a.getZoneID(ctx, domainName)
	if err != nil {
		return nil, err
	}

	page, err := a.client.DNS.Records.List(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(zoneID),
		Name:   cloudflare.F(dns.RecordListParamsName{Exact: cloudflare.F(domainName)}),
		Type:   cloudflare.F(dns.RecordListParamsType(recordType)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list dns records: %w", err)
	}

	result := make([]domain.DNSRecord, 0, len(page.Result))
	for _, r := range page.Result {
		result = append(result, domain.DNSRecord{
			ID:      r.ID,
			Name:    r.Name,
			Type:    string(r.Type),
			Content: r.Content,
			Proxied: r.Proxied,
		})
	}

	return result, nil
}

func (a *adapter) CreateRecord(ctx context.Context, domainName string, recordType string, content string, proxied bool) error {
	zoneID, err := a.getZoneID(ctx, domainName)
	if err != nil {
		return err
	}

	_, err = a.client.DNS.Records.New(ctx, dns.RecordNewParams{
		ZoneID: cloudflare.F(zoneID),
		Body: dns.RecordNewParamsBody{
			Name:    cloudflare.F(domainName),
			Type:    cloudflare.F(dns.RecordNewParamsBodyType(recordType)),
			Content: cloudflare.F(content),
			Proxied: cloudflare.F(proxied),
			TTL:     cloudflare.F(dns.TTL(dns.TTL1)),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create dns record: %w", err)
	}

	return nil
}

func (a *adapter) UpdateRecord(ctx context.Context, recordID string, domainName string, recordType string, content string, proxied bool) error {
	zoneID, err := a.getZoneID(ctx, domainName)
	if err != nil {
		return err
	}

	_, err = a.client.DNS.Records.Update(ctx, recordID, dns.RecordUpdateParams{
		ZoneID: cloudflare.F(zoneID),
		Body: dns.RecordUpdateParamsBody{
			Name:    cloudflare.F(domainName),
			Type:    cloudflare.F(dns.RecordUpdateParamsBodyType(recordType)),
			Content: cloudflare.F(content),
			Proxied: cloudflare.F(proxied),
			TTL:     cloudflare.F(dns.TTL(dns.TTL1)),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update dns record: %w", err)
	}

	return nil
}
