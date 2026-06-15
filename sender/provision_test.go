package main

import (
	"os"
	"path/filepath"
	"testing"
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
