package main

import "testing"

func TestResolveEndpointFromFlagWins(t *testing.T) {
	if got := resolveEndpointFrom("https://flag/v1/logs", "https://env/v1/logs", "https://file/v1/logs"); got != "https://flag/v1/logs" {
		t.Fatalf("got %q, want the flag value", got)
	}
}

func TestResolveEndpointFromEnvBeatsFile(t *testing.T) {
	if got := resolveEndpointFrom("", "https://env/v1/logs", "https://file/v1/logs"); got != "https://env/v1/logs" {
		t.Fatalf("got %q, want the env value over the file", got)
	}
}

func TestResolveEndpointFromFileFallback(t *testing.T) {
	if got := resolveEndpointFrom("", "", "https://file/v1/logs"); got != "https://file/v1/logs" {
		t.Fatalf("got %q, want the env-file fallback", got)
	}
}

func TestResolveEndpointFromNeitherIsEmpty(t *testing.T) {
	if got := resolveEndpointFrom("", "", ""); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}
