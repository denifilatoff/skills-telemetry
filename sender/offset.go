package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// OffsetStore persists a per-session byte offset into a harness transcript, so
// each Stop run ingests only the lines written since the previous run. It is
// harness-agnostic: callers namespace the key (for example "codex:<session>")
// to keep different harnesses from colliding. Named for the byte offset it
// holds, not the Cursor harness.
type OffsetStore struct {
	Dir string
}

// DefaultOffsetStore roots the offset directory in the user cache dir, beside
// the spool. It returns an error when no cache dir is available.
func DefaultOffsetStore() (*OffsetStore, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(base, pkgName, "offsets")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &OffsetStore{Dir: dir}, nil
}

// path maps a key to a safe file name, regardless of the key's shape.
func (o *OffsetStore) path(key string) string {
	safe := strings.NewReplacer("/", "_", "\\", "_", ":", "_", ".", "_").Replace(key)
	return filepath.Join(o.Dir, safe+".offset")
}

// Load returns the stored byte offset for key, or 0 when none is recorded or
// the stored value is unreadable.
func (o *OffsetStore) Load(key string) int64 {
	b, err := os.ReadFile(o.path(key))
	if err != nil {
		return 0
	}
	n, err := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// Save records the byte offset for key with an atomic temp-file rename.
func (o *OffsetStore) Save(key string, off int64) error {
	final := o.path(key)
	tmp := final + ".tmp"
	if err := os.WriteFile(tmp, []byte(strconv.FormatInt(off, 10)), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, final)
}
