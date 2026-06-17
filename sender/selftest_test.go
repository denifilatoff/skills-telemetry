package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSelftestDeliversProbeAndClearsIt(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := &Outbox{Dir: t.TempDir()}
	res, err := runSelftest(s, srv.URL, "", nil, 2*time.Second)
	if err != nil {
		t.Fatalf("selftest: %v", err)
	}
	if !res.Delivered {
		t.Fatal("want Delivered true on HTTP 200")
	}
	files, _ := s.List()
	if len(files) != 0 {
		t.Fatalf("probe should have left the outbox: %d remain", len(files))
	}
}

func TestSelftestKeepsProbeOnFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := &Outbox{Dir: t.TempDir()}
	res, err := runSelftest(s, srv.URL, "", nil, 2*time.Second)
	if err == nil {
		t.Fatal("want error when the collector rejects the probe")
	}
	if res.Delivered {
		t.Fatal("want Delivered false on failure")
	}
	if n := probesRemaining(s); n != 1 {
		t.Fatalf("probe should remain in the outbox: %d probes", n)
	}
}

func TestSelftestErrorsWhenUnprovisioned(t *testing.T) {
	s := &Outbox{Dir: t.TempDir()}
	if _, err := runSelftest(s, "", "", nil, time.Second); err == nil {
		t.Fatal("want error when no endpoint is configured")
	}
}
