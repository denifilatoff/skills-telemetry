package main

import "testing"

func TestOffsetStoreRoundTrip(t *testing.T) {
	o := &OffsetStore{Dir: t.TempDir()}
	if got := o.Load("codex:s1"); got != 0 {
		t.Fatalf("fresh offset = %d, want 0", got)
	}
	if err := o.Save("codex:s1", 4096); err != nil {
		t.Fatalf("save: %v", err)
	}
	if got := o.Load("codex:s1"); got != 4096 {
		t.Fatalf("offset = %d, want 4096", got)
	}
}

func TestOffsetStoreKeysAreIsolated(t *testing.T) {
	o := &OffsetStore{Dir: t.TempDir()}
	_ = o.Save("codex:s1", 10)
	_ = o.Save("codex:s2", 20)
	if o.Load("codex:s1") != 10 || o.Load("codex:s2") != 20 {
		t.Fatal("keys collided")
	}
}

func TestOffsetStoreOverwrite(t *testing.T) {
	o := &OffsetStore{Dir: t.TempDir()}
	_ = o.Save("k", 5)
	_ = o.Save("k", 9)
	if got := o.Load("k"); got != 9 {
		t.Fatalf("offset = %d, want 9", got)
	}
}
