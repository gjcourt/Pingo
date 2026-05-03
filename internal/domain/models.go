package domain

import "fmt"

// IPVersion represents the version of the IP address (IPv4 or IPv6).
type IPVersion string

const (
	IPv4 IPVersion = "IPv4"
	IPv6 IPVersion = "IPv6"
)

// RecordType returns the DNS record type for the IP version.
func (v IPVersion) RecordType() string {
	switch v {
	case IPv4:
		return "A"
	case IPv6:
		return "AAAA"
	default:
		return ""
	}
}

// DomainConfig represents the configuration for a specific domain to be updated.
type DomainConfig struct {
	Name    string
	IPType  IPVersion
	Proxied bool
}

// DNSRecord represents an existing DNS record in the provider.
type DNSRecord struct {
	ID      string
	Name    string
	Type    string
	Content string
	Proxied bool
}

// String provides a string representation of the DNSRecord.
func (r DNSRecord) String() string {
	return fmt.Sprintf("%s %s %s (Proxied: %t)", r.Name, r.Type, r.Content, r.Proxied)
}
