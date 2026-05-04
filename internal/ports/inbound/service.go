// Package inbound holds the driving-port interfaces: contracts that the app layer exposes.
package inbound

import (
	"context"

	"github.com/george/pingo/internal/domain"
)

// DDNSService is the driving port for the DDNS updater.
type DDNSService interface {
	UpdateDomains(ctx context.Context, configs []domain.DomainConfig) error
}
