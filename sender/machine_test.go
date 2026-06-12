package main

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

var uuidV4Re = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestResolveMachineIDMintsAndPersists(t *testing.T) {
	dir := t.TempDir()
	id := resolveMachineIDFrom(dir)
	if !uuidV4Re.MatchString(id) {
		t.Fatalf("not a v4 UUID: %q", id)
	}
	path := filepath.Join(dir, "qubership-skills-telemetry", "machine-id")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("id not persisted: %v", err)
	}
	if got := string(b); got != id+"\n" {
		t.Fatalf("file = %q, want %q", got, id+"\n")
	}
}

func TestResolveMachineIDStableAcrossCalls(t *testing.T) {
	dir := t.TempDir()
	first := resolveMachineIDFrom(dir)
	second := resolveMachineIDFrom(dir)
	if first != second {
		t.Fatalf("id changed between calls: %q vs %q", first, second)
	}
}
