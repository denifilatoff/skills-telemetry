package main

import "testing"

func TestResolveEndpointFromFlagWins(t *testing.T) {
	if got := resolveEndpointFrom("https://flag/v1/logs", "https://env/v1/logs"); got != "https://flag/v1/logs" {
		t.Fatalf("got %q, want the flag value", got)
	}
}

func TestResolveEndpointFromEnvFallback(t *testing.T) {
	if got := resolveEndpointFrom("", "https://env/v1/logs"); got != "https://env/v1/logs" {
		t.Fatalf("got %q, want the env fallback", got)
	}
}

func TestResolveEndpointFromNeitherIsEmpty(t *testing.T) {
	if got := resolveEndpointFrom("", ""); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}
