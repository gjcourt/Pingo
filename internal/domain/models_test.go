package domain_test

import (
	"testing"

	"github.com/george/pingo/internal/domain"
)

func TestIPVersion_RecordType(t *testing.T) {
	tests := []struct {
		name     string
		version  domain.IPVersion
		expected string
	}{
		{
			name:     "IPv4 returns A",
			version:  domain.IPv4,
			expected: "A",
		},
		{
			name:     "IPv6 returns AAAA",
			version:  domain.IPv6,
			expected: "AAAA",
		},
		{
			name:     "Unknown returns empty",
			version:  domain.IPVersion("Unknown"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.RecordType(); got != tt.expected {
				t.Errorf("IPVersion.RecordType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDNSRecord_String(t *testing.T) {
	record := domain.DNSRecord{
		ID:      "123",
		Name:    "example.com",
		Type:    "A",
		Content: "1.2.3.4",
		Proxied: true,
	}

	expected := "example.com A 1.2.3.4 (Proxied: true)"
	if got := record.String(); got != expected {
		t.Errorf("DNSRecord.String() = %v, want %v", got, expected)
	}
}
