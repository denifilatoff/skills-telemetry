package main

import "os"

// resolveEndpoint returns the OTLP/HTTP collector URL. Precedence: an explicit
// --endpoint= flag, then the SKILLS_TELEMETRY_ENDPOINT environment variable
// (for CI and automation overrides), then the provisioned env file so the
// binary is self-sufficient when invoked directly (not through the bootstrap).
// Empty when none is set — the flush then becomes a no-op.
func resolveEndpoint(flag string) string {
	return resolveEndpointFrom(flag, os.Getenv("SKILLS_TELEMETRY_ENDPOINT"), pkgEnv()["SKILLS_TELEMETRY_ENDPOINT"])
}

// resolveEndpointFrom is the testable core: flag wins, then env, then the
// env-file value.
func resolveEndpointFrom(flag, env, fileVal string) string {
	if flag != "" {
		return flag
	}
	if env != "" {
		return env
	}
	return fileVal
}
