package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEnvReadsKeyValueLines(t *testing.T) {
	in := []byte("SKILLS_TELEMETRY_ENDPOINT=https://otel.example/v1/logs\nSKILLS_TELEMETRY_TOKEN=abc123\n")
	got := parseEnv(in)
	if got["SKILLS_TELEMETRY_ENDPOINT"] != "https://otel.example/v1/logs" {
		t.Fatalf("endpoint = %q", got["SKILLS_TELEMETRY_ENDPOINT"])
	}
	if got["SKILLS_TELEMETRY_TOKEN"] != "abc123" {
		t.Fatalf("token = %q", got["SKILLS_TELEMETRY_TOKEN"])
	}
}

func TestParseEnvSkipsBlanksCommentsAndTrims(t *testing.T) {
	in := []byte("\n# a comment\n  SKILLS_TELEMETRY_ENDPOINT = https://x/v1/logs  \nnonsense-without-equals\n")
	got := parseEnv(in)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1: %v", len(got), got)
	}
	if got["SKILLS_TELEMETRY_ENDPOINT"] != "https://x/v1/logs" {
		t.Fatalf("endpoint = %q (key/value not trimmed)", got["SKILLS_TELEMETRY_ENDPOINT"])
	}
}

func TestLoadEnvFileMissingReturnsEmpty(t *testing.T) {
	got := loadEnvFile(filepath.Join(t.TempDir(), "qubership-skills-telemetry", "env"))
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0 for missing file", len(got))
	}
}

func TestLoadEnvFileReadsFromDisk(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "env")
	if err := os.WriteFile(p, []byte("SKILLS_TELEMETRY_ENDPOINT=https://disk/v1/logs\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got := loadEnvFile(p)
	if got["SKILLS_TELEMETRY_ENDPOINT"] != "https://disk/v1/logs" {
		t.Fatalf("endpoint = %q", got["SKILLS_TELEMETRY_ENDPOINT"])
	}
}
