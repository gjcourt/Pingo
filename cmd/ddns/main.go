package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/george/pingo/internal/adapters/driven/cloudflare"
	"github.com/george/pingo/internal/adapters/driven/ipfetcher"
	"github.com/george/pingo/internal/core/domain"
	"github.com/george/pingo/internal/core/services"
)

func main() {
	log.Println("Starting Cloudflare DDNS Updater...")

	// 1. Load Configuration from Environment Variables
	apiToken := os.Getenv("CF_API_TOKEN")
	if apiToken == "" {
		log.Fatal("CF_API_TOKEN environment variable is required")
	}

	domainsEnv := os.Getenv("DOMAINS")
	if domainsEnv == "" {
		log.Fatal("DOMAINS environment variable is required (comma-separated list of domains)")
	}

	proxiedEnv := os.Getenv("PROXIED")
	proxied := strings.ToLower(proxiedEnv) == "true" || proxiedEnv == "1"

	// Parse domains
	domainNames := strings.Split(domainsEnv, ",")
	var configs []domain.DomainConfig
	for _, name := range domainNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		// For simplicity, we configure both IPv4 and IPv6 for each domain.
		// In a more advanced setup, this could be configurable per domain.
		configs = append(configs, domain.DomainConfig{
			Name:    name,
			IPType:  domain.IPv4,
			Proxied: proxied,
		})
		configs = append(configs, domain.DomainConfig{
			Name:    name,
			IPType:  domain.IPv6,
			Proxied: proxied,
		})
	}

	if len(configs) == 0 {
		log.Fatal("No valid domains configured")
	}

	// 2. Initialize Adapters (Driven Ports)
	cfAdapter, err := cloudflare.NewAdapter(apiToken)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudflare adapter: %v", err)
	}

	ipFetcherAdapter := ipfetcher.NewCloudflareTraceFetcher()

	// 3. Initialize Application Service
	ddnsService := services.NewDDNSService(ipFetcherAdapter, cfAdapter)

	// 4. Execute the update
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = ddnsService.UpdateDomains(ctx, configs)
	if err != nil {
		log.Fatalf("Failed to update domains: %v", err)
	}

	log.Println("DDNS update completed successfully.")
}
