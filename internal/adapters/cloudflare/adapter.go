package cloudflare

import (
	"context"
	"fmt"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"
	"github.com/george/pingo/internal/domain"
	"github.com/george/pingo/internal/ports/outbound"
)

type adapter struct {
	api *cf.API
}

// NewAdapter creates a new Cloudflare DNS provider adapter.
func NewAdapter(apiToken string) (outbound.DNSProvider, error) {
	api, err := cf.NewWithAPIToken(apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudflare api: %w", err)
	}
	return &adapter{api: api}, nil
}

// getZoneID finds the Zone ID for a given domain name. It walks the labels
// from most-specific to least-specific (e.g. a.b.example.com → b.example.com
// → example.com) and returns the first matching zone. The context propagates
// to all underlying HTTP calls so cancellation/deadlines are honoured.
func (a *adapter) getZoneID(ctx context.Context, domainName string) (string, error) {
	parts := strings.Split(domainName, ".")
	var lastErr error
	for i := 0; i < len(parts)-1; i++ {
		zoneName := strings.Join(parts[i:], ".")
		res, err := a.api.ListZonesContext(ctx, cf.WithZoneFilters(zoneName, "", ""))
		if err != nil {
			lastErr = err
			continue
		}
		for _, z := range res.Result {
			if z.Name == zoneName && z.ID != "" {
				return z.ID, nil
			}
		}
	}
	if lastErr != nil {
		return "", fmt.Errorf("could not find zone for domain %s: %w", domainName, lastErr)
	}
	return "", fmt.Errorf("could not find zone for domain: %s", domainName)
}

func (a *adapter) GetRecords(ctx context.Context, domainName string, recordType string) ([]domain.DNSRecord, error) {
	zoneID, err := a.getZoneID(ctx, domainName)
	if err != nil {
		return nil, err
	}

	rc := cf.ZoneIdentifier(zoneID)
	params := cf.ListDNSRecordsParams{
		Name: domainName,
		Type: recordType,
	}

	records, _, err := a.api.ListDNSRecords(ctx, rc, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list dns records: %w", err)
	}

	result := make([]domain.DNSRecord, 0, len(records))
	for _, r := range records {
		proxied := false
		if r.Proxied != nil {
			proxied = *r.Proxied
		}
		result = append(result, domain.DNSRecord{
			ID:      r.ID,
			Name:    r.Name,
			Type:    r.Type,
			Content: r.Content,
			Proxied: proxied,
		})
	}

	return result, nil
}

func (a *adapter) CreateRecord(ctx context.Context, domainName string, recordType string, content string, proxied bool) error {
	zoneID, err := a.getZoneID(ctx, domainName)
	if err != nil {
		return err
	}

	rc := cf.ZoneIdentifier(zoneID)
	params := cf.CreateDNSRecordParams{
		Name:    domainName,
		Type:    recordType,
		Content: content,
		Proxied: &proxied,
		TTL:     1, // 1 = Automatic
	}

	_, err = a.api.CreateDNSRecord(ctx, rc, params)
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

	rc := cf.ZoneIdentifier(zoneID)
	params := cf.UpdateDNSRecordParams{
		ID:      recordID,
		Name:    domainName,
		Type:    recordType,
		Content: content,
		Proxied: &proxied,
		TTL:     1, // 1 = Automatic
	}

	_, err = a.api.UpdateDNSRecord(ctx, rc, params)
	if err != nil {
		return fmt.Errorf("failed to update dns record: %w", err)
	}

	return nil
}
