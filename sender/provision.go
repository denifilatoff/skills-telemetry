package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// applyProvision is the deterministic core the skill and the one-liner both
// call: it writes only the fields it is given. The endpoint and token go into
// the env file (merged, so they can be set in separate runs); a CA path is
// validated and copied to ca.crt. Empty fields are left untouched, which keeps
// re-running provision safe.
func applyProvision(configDir, endpoint, caPath, token string) error {
	updates := map[string]string{}
	if endpoint != "" {
		updates["SKILLS_TELEMETRY_ENDPOINT"] = endpoint
	}
	if token != "" {
		updates["SKILLS_TELEMETRY_TOKEN"] = token
	}
	if len(updates) > 0 {
		if err := writeEnvFile(configDir, updates); err != nil {
			return err
		}
	}
	if caPath != "" {
		if err := copyCAFile(configDir, caPath); err != nil {
			return err
		}
	}
	return nil
}

// writeEnvFile merges updates into the env file under configDir and writes it
// back atomically (temp file + rename) with 0600 permissions, since the file
// may hold the token. Existing keys not in updates are preserved, so callers
// can set the endpoint and the token in separate steps. Sorted output makes the
// write idempotent for an unchanged set of values.
func writeEnvFile(configDir string, updates map[string]string) error {
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(configDir, "env")

	merged := loadEnvFile(path)
	for k, v := range updates {
		merged[k] = v
	}

	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf []byte
	for _, k := range keys {
		buf = append(buf, fmt.Sprintf("%s=%s\n", k, merged[k])...)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
