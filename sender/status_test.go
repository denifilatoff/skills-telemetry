package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGatherStatusReportsProvisionedState(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), pkgName)
	if err := os.WriteFile(filepath.Join(t.TempDir(), "src.crt"), selfSignedPEM(t), 0o644); err != nil {
		t.Fatal(err)
	}
	// place a ca.crt via the real writer so the test mirrors provisioning
	src := filepath.Join(t.TempDir(), "src.crt")
	_ = os.WriteFile(src, selfSignedPEM(t), 0o644)
	if err := copyCAFile(cfg, src); err != nil {
		t.Fatal(err)
	}

	s := &Outbox{Dir: t.TempDir()}
	seed(t, s, 2)

	r := gatherStatus(s, cfg, "https://otel.example/v1/logs")
	if !r.Provisioned {
		t.Fatal("want provisioned when an endpoint is set")
	}
	if !r.CAFound {
		t.Fatal("want CAFound when ca.crt is present")
	}
	if r.Buffered != 2 {
		t.Fatalf("buffered = %d, want 2", r.Buffered)
	}
	if r.Endpoint != "https://otel.example/v1/logs" {
		t.Fatalf("endpoint = %q", r.Endpoint)
	}
}

func TestGatherStatusUnprovisionedWhenNoEndpoint(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), pkgName)
	s := &Outbox{Dir: t.TempDir()}
	r := gatherStatus(s, cfg, "")
	if r.Provisioned {
		t.Fatal("want not provisioned when endpoint is empty")
	}
	if r.CAFound {
		t.Fatal("want CAFound false when no ca.crt")
	}
}

func TestFormatStatusFlagsNextStepWhenUnprovisioned(t *testing.T) {
	out := formatStatus(statusReport{Provisioned: false, ConfigDir: "/cfg"})
	if !strings.Contains(strings.ToLower(out), "not provisioned") {
		t.Fatalf("output should flag the unprovisioned state, got:\n%s", out)
	}
}
