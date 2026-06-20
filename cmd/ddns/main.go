package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/george/pingo/internal/adapters/adguard"
	"github.com/george/pingo/internal/adapters/cloudflare"
	"github.com/george/pingo/internal/adapters/ipfetcher"
	"github.com/george/pingo/internal/app"
	"github.com/george/pingo/internal/domain"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	logger.Info("starting Cloudflare DDNS updater")

	// 1. Load Configuration from Environment Variables
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	if apiToken == "" {
		logger.Error("CLOUDFLARE_API_TOKEN environment variable is required")
		os.Exit(1)
	}

	domainsEnv := os.Getenv("DOMAINS")
	if domainsEnv == "" {
		logger.Error("DOMAINS environment variable is required (comma-separated list of domains)")
		os.Exit(1)
	}

	proxiedEnv := os.Getenv("PROXIED")
	proxied := strings.ToLower(proxiedEnv) == "true" || proxiedEnv == "1"

	// Parse domains
	domainNames := strings.Split(domainsEnv, ",")
	configs := make([]domain.DomainConfig, 0, len(domainNames))
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
		logger.Error("no valid domains configured")
		os.Exit(1)
	}

	// 2. Initialize Adapters (Driven Ports)
	cfAdapter, err := cloudflare.NewAdapter(apiToken)
	if err != nil {
		logger.Error("failed to initialize cloudflare adapter", "err", err)
		os.Exit(1)
	}

	providers := []app.NamedProvider{{Name: "cloudflare", Provider: cfAdapter}}

	// Optionally sync the same records into one or more AdGuard Home instances
	// as DNS rewrites. This keeps a split-horizon override (e.g. a host caught
	// by an internal *.domain wildcard) pointed at the live public IP for LAN
	// clients. Enabled only when ADGUARD_URLS is set; comma-separated to cover
	// an HA pair of independent AdGuard instances.
	adguardURLs := os.Getenv("ADGUARD_URLS")
	if adguardURLs != "" {
		agUser := os.Getenv("ADGUARD_USERNAME")
		agPass := os.Getenv("ADGUARD_PASSWORD")
		for _, rawURL := range strings.Split(adguardURLs, ",") {
			rawURL = strings.TrimSpace(rawURL)
			if rawURL == "" {
				continue
			}
			agAdapter, agErr := adguard.NewAdapter(rawURL, agUser, agPass, nil)
			if agErr != nil {
				logger.Error("failed to initialize adguard adapter", "url", rawURL, "err", agErr)
				os.Exit(1)
			}
			providers = append(providers, app.NamedProvider{Name: "adguard(" + rawURL + ")", Provider: agAdapter})
		}
	}

	ipFetcherAdapter := ipfetcher.NewCloudflareTraceFetcher()

	// 3. Initialize Application Service
	ddnsService := app.NewDDNSServiceMulti(ipFetcherAdapter, providers, logger)

	// 4. Execute the update
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ddnsService.UpdateDomains(ctx, configs); err != nil {
		logger.Error("failed to update domains", "err", err)
		os.Exit(1)
	}

	logger.Info("DDNS update completed successfully")
}
