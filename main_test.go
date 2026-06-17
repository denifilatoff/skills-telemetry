package main

import "testing"

func TestRunVersion(t *testing.T) {
	var out string
	code := run([]string{"version"}, func(s string) { out = s })
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if out != version+"\n" {
		t.Fatalf("output = %q, want %q", out, version+"\n")
	}
}

func TestRunUnknownCommand(t *testing.T) {
	code := run([]string{"bogus"}, func(string) {})
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}
