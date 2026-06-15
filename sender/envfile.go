package main

import (
	"os"
	"strings"
)

// parseEnv reads KEY=VALUE lines into a map. Blank lines and lines starting
// with '#' are ignored, as are lines without '='. Keys and values are trimmed
// of surrounding whitespace. This is the format the provisioned env file uses,
// shared with the shell bootstrap so there is one config format per machine.
func parseEnv(b []byte) map[string]string {
	out := make(map[string]string)
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return out
}

// loadEnvFile parses the env file at path. A missing or unreadable file yields
// an empty map, not an error: an unprovisioned machine is a valid state.
func loadEnvFile(path string) map[string]string {
	b, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}
	}
	return parseEnv(b)
}
