package main

import (
	"os"
	"path/filepath"
	"strings"
)

// resolveToken returns the collector bearer token. It prefers the
// SKILLS_TELEMETRY_TOKEN environment variable and falls back to a per-user
// secret file at <UserConfigDir>/qubership-skills-telemetry/token. Empty when
// neither is set — the flush then sends no Authorization header.
func resolveToken() string {
	dir, _ := os.UserConfigDir()
	return resolveTokenFrom(os.Getenv("SKILLS_TELEMETRY_TOKEN"), dir)
}

// resolveTokenFrom is the testable core: env wins; otherwise read the secret
// file under configDir. Surrounding whitespace/newlines are trimmed.
func resolveTokenFrom(env, configDir string) string {
	if env != "" {
		return env
	}
	if configDir == "" {
		return ""
	}
	b, err := os.ReadFile(filepath.Join(configDir, "qubership-skills-telemetry", "token"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
