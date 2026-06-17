package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteEnvFileCreatesWithSecurePerms(t *testing.T) {
	dir := filepath.Join(t.TempDir(), pkgName)
	if err := writeEnvFile(dir, map[string]string{"SKILLS_TELEMETRY_ENDPOINT": "https://x/v1/logs"}); err != nil {
		t.Fatal(err)
	}
	got := loadEnvFile(filepath.Join(dir, "env"))
	if got["SKILLS_TELEMETRY_ENDPOINT"] != "https://x/v1/logs" {
		t.Fatalf("endpoint = %q", got["SKILLS_TELEMETRY_ENDPOINT"])
	}
	fi, err := os.Stat(filepath.Join(dir, "env"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Fatalf("perm = %o, want 600 (the file may hold a token)", perm)
	}
}

func TestWriteEnvFileMergesKeepingExistingKeys(t *testing.T) {
	dir := filepath.Join(t.TempDir(), pkgName)
	if err := writeEnvFile(dir, map[string]string{"SKILLS_TELEMETRY_ENDPOINT": "https://x/v1/logs"}); err != nil {
		t.Fatal(err)
	}
	if err := writeEnvFile(dir, map[string]string{"SKILLS_TELEMETRY_TOKEN": "secret"}); err != nil {
		t.Fatal(err)
	}
	got := loadEnvFile(filepath.Join(dir, "env"))
	if got["SKILLS_TELEMETRY_ENDPOINT"] != "https://x/v1/logs" || got["SKILLS_TELEMETRY_TOKEN"] != "secret" {
		t.Fatalf("merge lost a key: %v", got)
	}
}

func TestWriteEnvFileOverwritesProvidedKey(t *testing.T) {
	dir := filepath.Join(t.TempDir(), pkgName)
	_ = writeEnvFile(dir, map[string]string{"SKILLS_TELEMETRY_ENDPOINT": "https://old/v1/logs"})
	_ = writeEnvFile(dir, map[string]string{"SKILLS_TELEMETRY_ENDPOINT": "https://new/v1/logs"})
	got := loadEnvFile(filepath.Join(dir, "env"))
	if got["SKILLS_TELEMETRY_ENDPOINT"] != "https://new/v1/logs" {
		t.Fatalf("endpoint = %q, want overwritten", got["SKILLS_TELEMETRY_ENDPOINT"])
	}
}

func TestApplyProvisionWritesEndpointTokenAndCA(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), pkgName)
	src := filepath.Join(t.TempDir(), "src.crt")
	if err := os.WriteFile(src, selfSignedPEM(t), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := applyProvision(cfg, "https://otel.example/v1/logs", src, "secret"); err != nil {
		t.Fatal(err)
	}
	env := loadEnvFile(filepath.Join(cfg, "env"))
	if env["SKILLS_TELEMETRY_ENDPOINT"] != "https://otel.example/v1/logs" {
		t.Fatalf("endpoint = %q", env["SKILLS_TELEMETRY_ENDPOINT"])
	}
	if env["SKILLS_TELEMETRY_TOKEN"] != "secret" {
		t.Fatalf("token = %q", env["SKILLS_TELEMETRY_TOKEN"])
	}
	if _, err := os.Stat(filepath.Join(cfg, caFileName)); err != nil {
		t.Fatalf("ca.crt not written: %v", err)
	}
}

func TestApplyProvisionOnlyWritesProvidedFields(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), pkgName)
	if err := applyProvision(cfg, "https://otel.example/v1/logs", "", ""); err != nil {
		t.Fatal(err)
	}
	env := loadEnvFile(filepath.Join(cfg, "env"))
	if _, ok := env["SKILLS_TELEMETRY_TOKEN"]; ok {
		t.Fatal("token key should be absent when no token was given")
	}
	if _, err := os.Stat(filepath.Join(cfg, caFileName)); err == nil {
		t.Fatal("ca.crt should not exist when no CA path was given")
	}
}

func TestWriteEnvFileIsIdempotent(t *testing.T) {
	dir := filepath.Join(t.TempDir(), pkgName)
	kv := map[string]string{"SKILLS_TELEMETRY_ENDPOINT": "https://x/v1/logs", "SKILLS_TELEMETRY_TOKEN": "t"}
	_ = writeEnvFile(dir, kv)
	first, _ := os.ReadFile(filepath.Join(dir, "env"))
	_ = writeEnvFile(dir, kv)
	second, _ := os.ReadFile(filepath.Join(dir, "env"))
	if string(first) != string(second) {
		t.Fatalf("not idempotent:\n%q\n%q", first, second)
	}
}

func TestParseProvisionFlags(t *testing.T) {
	endpoint, ca := parseProvisionFlags([]string{"--endpoint=https://x/v1/logs", "--ca=/tmp/ca.crt"})
	if endpoint != "https://x/v1/logs" {
		t.Fatalf("endpoint = %q", endpoint)
	}
	if ca != "/tmp/ca.crt" {
		t.Fatalf("ca = %q", ca)
	}
}

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
