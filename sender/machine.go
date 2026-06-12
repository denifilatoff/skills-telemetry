package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// resolveMachineID returns a stable, anonymous identifier for this install.
// It is a random UUID generated once and persisted under the per-user config
// dir, never derived from hardware or the username. The collector uses it only
// to tell installs apart (e.g. to spot one install skewing skill-usage counts),
// not to identify the user. Returns "" if no config dir is available.
func resolveMachineID() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return ""
	}
	return resolveMachineIDFrom(dir)
}

// resolveMachineIDFrom is the testable core: read the id file under configDir,
// or create it atomically on first run. O_EXCL guards against two concurrent
// processes each minting a different id — the loser re-reads the winner's file.
func resolveMachineIDFrom(configDir string) string {
	pkgDir := filepath.Join(configDir, "qubership-skills-telemetry")
	path := filepath.Join(pkgDir, "machine-id")
	if b, err := os.ReadFile(path); err == nil {
		if id := strings.TrimSpace(string(b)); id != "" {
			return id
		}
	}
	id := newUUID()
	if id == "" {
		return ""
	}
	if err := os.MkdirAll(pkgDir, 0o700); err != nil {
		return id // usable this run, just not persisted
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		// Lost the race or cannot create: prefer an already-written id.
		if b, rerr := os.ReadFile(path); rerr == nil {
			if existing := strings.TrimSpace(string(b)); existing != "" {
				return existing
			}
		}
		return id
	}
	_, _ = f.WriteString(id + "\n")
	_ = f.Close()
	return id
}

// newUUID returns a random RFC 4122 version 4 UUID, or "" if the system CSPRNG
// is unavailable.
func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ""
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
