// Package adguard implements the outbound.DNSProvider port against an
// AdGuard Home instance's DNS-rewrite API (/control/rewrite/*).
//
// AdGuard rewrites override what a name resolves to for clients of that
// AdGuard instance. Pingo uses this to keep a split-horizon override in sync
// with the public IP: when a wildcard like *.example.com points LAN clients at
// an internal reverse proxy, a more-specific rewrite for a single host (e.g. a
// WireGuard endpoint) must instead return the current public IP so the handshake
// still works from inside the LAN (via NAT hairpin).
//
// The rewrite model has no record IDs and no update verb — only list, add and
// delete. We map it onto outbound.DNSProvider by treating the stored answer as
// the record ID, so UpdateRecord is delete(old answer)+add(new answer).
package adguard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/george/pingo/internal/domain"
	"github.com/george/pingo/internal/ports/outbound"
)

const defaultTimeout = 10 * time.Second

// rewrite is the wire shape of an AdGuard Home DNS-rewrite entry.
type rewrite struct {
	Domain string `json:"domain"`
	Answer string `json:"answer"`
}

type adapter struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

// NewAdapter creates an AdGuard Home DNS-rewrite provider. baseURL is the
// instance root (e.g. "http://10.0.0.2"); username/password are its HTTP basic
// auth credentials. A nil httpClient uses a default with a sane timeout.
func NewAdapter(baseURL, username, password string, httpClient *http.Client) (outbound.DNSProvider, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("adguard: base URL is required")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}
	return &adapter{
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		password: password,
		client:   httpClient,
	}, nil
}

// answerMatchesType reports whether an AdGuard answer (an IP literal) belongs to
// the given DNS record type, so an A reconcile never disturbs an AAAA rewrite.
func answerMatchesType(answer, recordType string) bool {
	ip := net.ParseIP(answer)
	if ip == nil {
		return false
	}
	switch recordType {
	case "A":
		return ip.To4() != nil
	case "AAAA":
		return ip.To4() == nil && ip.To16() != nil
	default:
		return false
	}
}

func (a *adapter) GetRecords(ctx context.Context, domainName string, recordType string) ([]domain.DNSRecord, error) {
	var entries []rewrite
	if err := a.do(ctx, http.MethodGet, "/control/rewrite/list", nil, &entries); err != nil {
		return nil, fmt.Errorf("adguard: list rewrites: %w", err)
	}

	records := make([]domain.DNSRecord, 0, 1)
	for _, e := range entries {
		if e.Domain == domainName && answerMatchesType(e.Answer, recordType) {
			records = append(records, domain.DNSRecord{
				// The answer is the natural identity of a rewrite: it is what
				// /control/rewrite/delete matches on.
				ID:      e.Answer,
				Name:    e.Domain,
				Type:    recordType,
				Content: e.Answer,
				Proxied: false, // AdGuard rewrites have no proxy concept.
			})
		}
	}
	return records, nil
}

func (a *adapter) CreateRecord(ctx context.Context, domainName string, recordType string, content string, _ bool) error {
	if err := a.do(ctx, http.MethodPost, "/control/rewrite/add", rewrite{Domain: domainName, Answer: content}, nil); err != nil {
		return fmt.Errorf("adguard: add rewrite %s -> %s: %w", domainName, content, err)
	}
	return nil
}

// UpdateRecord changes a rewrite's answer. AdGuard has no update verb, so this
// removes the old answer (carried in recordID) and adds the new one. If the add
// fails after the delete succeeds, the rewrite is left absent and the next run
// recreates it via CreateRecord.
func (a *adapter) UpdateRecord(ctx context.Context, recordID string, domainName string, _ string, content string, _ bool) error {
	if recordID == content {
		return nil
	}
	if err := a.do(ctx, http.MethodPost, "/control/rewrite/delete", rewrite{Domain: domainName, Answer: recordID}, nil); err != nil {
		return fmt.Errorf("adguard: delete rewrite %s -> %s: %w", domainName, recordID, err)
	}
	if err := a.do(ctx, http.MethodPost, "/control/rewrite/add", rewrite{Domain: domainName, Answer: content}, nil); err != nil {
		return fmt.Errorf("adguard: re-add rewrite %s -> %s: %w", domainName, content, err)
	}
	return nil
}

// do executes a control-API call, encoding body (if non-nil) as JSON and
// decoding the response into out (if non-nil).
func (a *adapter) do(ctx context.Context, method, path string, body, out any) error {
	var reader *bytes.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(payload)
	} else {
		reader = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if a.username != "" || a.password != "" {
		req.SetBasicAuth(a.username, a.password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d from %s %s", resp.StatusCode, method, path)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
