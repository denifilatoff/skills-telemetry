package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveTokenFromEnvWins(t *testing.T) {
	if got := resolveTokenFrom("env-token", t.TempDir()); got != "env-token" {
		t.Fatalf("got %q, want env-token", got)
	}
}

func TestResolveTokenFromFileFallback(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "qubership-skills-telemetry")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "token"), []byte("  file-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := resolveTokenFrom("", dir); got != "file-token" {
		t.Fatalf("got %q, want file-token (trimmed)", got)
	}
}

func TestResolveTokenFromNeitherIsEmpty(t *testing.T) {
	if got := resolveTokenFrom("", t.TempDir()); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}
