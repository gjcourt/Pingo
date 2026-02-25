package ipfetcher_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/george/pingo/internal/adapters/driven/ipfetcher"
)

func TestCloudflareTraceFetcher_GetIPv4(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fl=114f114\nh=1.1.1.1\nip=203.0.113.1\nts=1678886400\nvisit_scheme=https\nuag=curl/7.81.0\ncolo=SJC\nsliver=none\nhttp=http/2\nloc=US\ntls=TLSv1.3\nsni=plaintext\nwarp=off\ngateway=off\nrbi=off\nkex=X25519\n"))
	}))
	defer server.Close()

	fetcher := ipfetcher.NewCloudflareTraceFetcherWithClient(server.Client(), server.URL, server.URL)

	ip, err := fetcher.GetIPv4(context.Background())
	if err != nil {
		t.Fatalf("GetIPv4() error = %v", err)
	}

	if ip != "203.0.113.1" {
		t.Errorf("GetIPv4() = %v, want %v", ip, "203.0.113.1")
	}
}

func TestCloudflareTraceFetcher_GetIPv6(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fl=114f114\nh=1.1.1.1\nip=2001:db8::1\nts=1678886400\nvisit_scheme=https\nuag=curl/7.81.0\ncolo=SJC\nsliver=none\nhttp=http/2\nloc=US\ntls=TLSv1.3\nsni=plaintext\nwarp=off\ngateway=off\nrbi=off\nkex=X25519\n"))
	}))
	defer server.Close()

	fetcher := ipfetcher.NewCloudflareTraceFetcherWithClient(server.Client(), server.URL, server.URL)

	ip, err := fetcher.GetIPv6(context.Background())
	if err != nil {
		t.Fatalf("GetIPv6() error = %v", err)
	}

	if ip != "2001:db8::1" {
		t.Errorf("GetIPv6() = %v, want %v", ip, "2001:db8::1")
	}
}

func TestCloudflareTraceFetcher_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	fetcher := ipfetcher.NewCloudflareTraceFetcherWithClient(server.Client(), server.URL, server.URL)

	_, err := fetcher.GetIPv4(context.Background())
	if err == nil {
		t.Fatal("GetIPv4() expected error, got nil")
	}
}
