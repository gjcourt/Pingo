package adguard_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/george/pingo/internal/adapters/adguard"
)

// fakeAdGuard is an in-memory stand-in for AdGuard Home's rewrite API.
type fakeAdGuard struct {
	mu       sync.Mutex
	rewrites []map[string]string
	gotAuth  string
	failAdd  bool // when set, /control/rewrite/add returns 500
}

func (f *fakeAdGuard) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/control/rewrite/list", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if u, p, ok := r.BasicAuth(); ok {
			f.gotAuth = u + ":" + p
		}
		_ = json.NewEncoder(w).Encode(f.rewrites)
	})
	mux.HandleFunc("/control/rewrite/add", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		fail := f.failAdd
		f.mu.Unlock()
		if fail {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		var e map[string]string
		_ = json.NewDecoder(r.Body).Decode(&e)
		f.mu.Lock()
		f.rewrites = append(f.rewrites, e)
		f.mu.Unlock()
	})
	mux.HandleFunc("/control/rewrite/delete", func(w http.ResponseWriter, r *http.Request) {
		var e map[string]string
		_ = json.NewDecoder(r.Body).Decode(&e)
		f.mu.Lock()
		kept := f.rewrites[:0]
		for _, x := range f.rewrites {
			if x["domain"] != e["domain"] || x["answer"] != e["answer"] {
				kept = append(kept, x)
			}
		}
		f.rewrites = kept
		f.mu.Unlock()
	})
	return mux
}

func TestAdGuardAdapter_CreatesMissingRewrite(t *testing.T) {
	fake := &fakeAdGuard{}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	a, err := adguard.NewAdapter(srv.URL, "admin", "secret", srv.Client())
	if err != nil {
		t.Fatalf("NewAdapter: %v", err)
	}

	got, err := a.GetRecords(context.Background(), "vpn.example.com", "A")
	if err != nil {
		t.Fatalf("GetRecords: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no records, got %d", len(got))
	}

	if err = a.CreateRecord(context.Background(), "vpn.example.com", "A", "1.2.3.4", false); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	got, err = a.GetRecords(context.Background(), "vpn.example.com", "A")
	if err != nil {
		t.Fatalf("GetRecords after create: %v", err)
	}
	if len(got) != 1 || got[0].Content != "1.2.3.4" {
		t.Fatalf("expected rewrite -> 1.2.3.4, got %+v", got)
	}
	if fake.gotAuth != "admin:secret" {
		t.Errorf("basic auth not sent correctly, got %q", fake.gotAuth)
	}
}

func TestAdGuardAdapter_UpdateIsDeleteThenAdd(t *testing.T) {
	fake := &fakeAdGuard{rewrites: []map[string]string{
		{"domain": "vpn.example.com", "answer": "1.1.1.1"},
	}}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	a, _ := adguard.NewAdapter(srv.URL, "", "", srv.Client())

	// The stored answer is the record ID we delete on.
	if err := a.UpdateRecord(context.Background(), "1.1.1.1", "vpn.example.com", "A", "2.2.2.2", false); err != nil {
		t.Fatalf("UpdateRecord: %v", err)
	}

	got, _ := a.GetRecords(context.Background(), "vpn.example.com", "A")
	if len(got) != 1 || got[0].Content != "2.2.2.2" {
		t.Fatalf("expected single rewrite -> 2.2.2.2, got %+v", got)
	}
}

func TestAdGuardAdapter_UpdateAddFailureLeavesRewriteAbsent(t *testing.T) {
	// delete succeeds, add fails -> UpdateRecord errors and the rewrite is
	// gone, so the next reconcile recreates it via CreateRecord.
	fake := &fakeAdGuard{
		rewrites: []map[string]string{{"domain": "vpn.example.com", "answer": "1.1.1.1"}},
		failAdd:  true,
	}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	a, _ := adguard.NewAdapter(srv.URL, "", "", srv.Client())

	err := a.UpdateRecord(context.Background(), "1.1.1.1", "vpn.example.com", "A", "2.2.2.2", false)
	if err == nil {
		t.Fatal("expected UpdateRecord to error when add fails")
	}

	got, _ := a.GetRecords(context.Background(), "vpn.example.com", "A")
	if len(got) != 0 {
		t.Fatalf("expected rewrite to be absent after failed re-add, got %+v", got)
	}

	// Recovery: with add working again, CreateRecord restores it.
	fake.mu.Lock()
	fake.failAdd = false
	fake.mu.Unlock()
	if err := a.CreateRecord(context.Background(), "vpn.example.com", "A", "2.2.2.2", false); err != nil {
		t.Fatalf("CreateRecord recovery: %v", err)
	}
	got, _ = a.GetRecords(context.Background(), "vpn.example.com", "A")
	if len(got) != 1 || got[0].Content != "2.2.2.2" {
		t.Fatalf("expected recovered rewrite -> 2.2.2.2, got %+v", got)
	}
}

func TestAdGuardAdapter_RecordTypeFiltering(t *testing.T) {
	fake := &fakeAdGuard{rewrites: []map[string]string{
		{"domain": "vpn.example.com", "answer": "1.2.3.4"},          // A
		{"domain": "vpn.example.com", "answer": "2606:4700:4700::"}, // AAAA
		{"domain": "other.example.com", "answer": "9.9.9.9"},        // different host
	}}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	a, _ := adguard.NewAdapter(srv.URL, "", "", srv.Client())

	v4, _ := a.GetRecords(context.Background(), "vpn.example.com", "A")
	if len(v4) != 1 || v4[0].Content != "1.2.3.4" {
		t.Fatalf("A query should return only the IPv4 rewrite, got %+v", v4)
	}
	v6, _ := a.GetRecords(context.Background(), "vpn.example.com", "AAAA")
	if len(v6) != 1 || v6[0].Content != "2606:4700:4700::" {
		t.Fatalf("AAAA query should return only the IPv6 rewrite, got %+v", v6)
	}
}
