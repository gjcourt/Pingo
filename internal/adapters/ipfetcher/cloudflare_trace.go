package ipfetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/george/pingo/internal/ports/outbound"
)

const (
	traceURLv4 = "https://1.1.1.1/cdn-cgi/trace"
	traceURLv6 = "https://[2606:4700:4700::1111]/cdn-cgi/trace"
)

type cloudflareTraceFetcher struct {
	client *http.Client
	urlV4  string
	urlV6  string
}

// NewCloudflareTraceFetcher creates a new IPFetcher using Cloudflare's trace endpoint.
func NewCloudflareTraceFetcher() outbound.IPFetcher {
	return &cloudflareTraceFetcher{
		client: &http.Client{},
		urlV4:  traceURLv4,
		urlV6:  traceURLv6,
	}
}

// NewCloudflareTraceFetcherWithClient creates a new IPFetcher with a custom HTTP client and URLs (useful for testing).
func NewCloudflareTraceFetcherWithClient(client *http.Client, urlV4, urlV6 string) outbound.IPFetcher {
	return &cloudflareTraceFetcher{
		client: client,
		urlV4:  urlV4,
		urlV6:  urlV6,
	}
}

func (f *cloudflareTraceFetcher) GetIPv4(ctx context.Context) (string, error) {
	return f.fetchIP(ctx, f.urlV4)
}

func (f *cloudflareTraceFetcher) GetIPv6(ctx context.Context) (string, error) {
	return f.fetchIP(ctx, f.urlV6)
}

func (f *cloudflareTraceFetcher) fetchIP(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the trace output to find the "ip=" line
	lines := strings.Split(string(bodyBytes), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ip=") {
			return strings.TrimPrefix(line, "ip="), nil
		}
	}

	return "", fmt.Errorf("ip address not found in trace output")
}
