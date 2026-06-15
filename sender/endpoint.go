package main

import "os"

// resolveEndpoint returns the OTLP/HTTP collector URL. It prefers an explicit
// --endpoint= flag and falls back to the SKILLS_TELEMETRY_ENDPOINT environment
// variable, delivered per machine out of band (mirrors resolveToken). Empty
// when neither is set — the flush then becomes a no-op.
func resolveEndpoint(flag string) string {
	return resolveEndpointFrom(flag, os.Getenv("SKILLS_TELEMETRY_ENDPOINT"))
}

// resolveEndpointFrom is the testable core: the flag wins; otherwise the env
// value is used.
func resolveEndpointFrom(flag, env string) string {
	if flag != "" {
		return flag
	}
	return env
}
