package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const markerName = ".lastflush"

// Spool is a machine-global directory holding one JSON file per buffered event.
type Spool struct {
	Dir string
}

// DefaultSpool returns the per-machine spool rooted in the user cache dir.
func DefaultSpool() (*Spool, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(base, "qubership-skills-telemetry", "spool")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Spool{Dir: dir}, nil
}

// Enqueue writes one event atomically (temp file + rename).
func (s *Spool) Enqueue(ev SkillEvent) error {
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%020d-%d-%s.json", time.Now().UnixNano(), os.Getpid(), randHex())
	final := filepath.Join(s.Dir, name)
	tmp := final + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, final)
}

// List returns event file names (not paths), oldest first, excluding temp files
// and the throttle marker.
func (s *Spool) List() ([]string, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || strings.HasPrefix(n, ".") || strings.HasSuffix(n, ".tmp") {
			continue
		}
		names = append(names, n)
	}
	sort.Strings(names) // filenames start with zero-padded nanos => chronological
	return names, nil
}

// Read decodes one event file by name.
func (s *Spool) Read(name string) (SkillEvent, error) {
	var ev SkillEvent
	b, err := os.ReadFile(filepath.Join(s.Dir, name))
	if err != nil {
		return ev, err
	}
	err = json.Unmarshal(b, &ev)
	return ev, err
}

// Remove deletes one event file by name.
func (s *Spool) Remove(name string) error {
	return os.Remove(filepath.Join(s.Dir, name))
}

// Rotate deletes the oldest event files until at most limit remain.
// Returns how many were dropped. (limit avoids shadowing the builtin cap/max.)
func (s *Spool) Rotate(limit int) (int, error) {
	names, err := s.List()
	if err != nil {
		return 0, err
	}
	if len(names) <= limit {
		return 0, nil
	}
	drop := names[:len(names)-limit] // List is oldest-first
	for _, n := range drop {
		if err := s.Remove(n); err != nil {
			return 0, err
		}
	}
	return len(drop), nil
}

func randHex() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
