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

// resolveTokenFrom is the testable core. Precedence: the env var, then the
// provisioned env file (the single config file provision writes), then the
// legacy standalone token file. Surrounding whitespace/newlines are trimmed.
func resolveTokenFrom(env, configDir string) string {
	if env != "" {
		return env
	}
	if configDir == "" {
		return ""
	}
	pkgDir := filepath.Join(configDir, pkgName)
	if tok := loadEnvFile(filepath.Join(pkgDir, "env"))["SKILLS_TELEMETRY_TOKEN"]; tok != "" {
		return tok
	}
	b, err := os.ReadFile(filepath.Join(pkgDir, "token"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
