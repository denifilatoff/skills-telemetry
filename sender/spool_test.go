package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSpoolEnqueueAndList(t *testing.T) {
	dir := t.TempDir()
	s := &Spool{Dir: dir}

	ev := SkillEvent{Agent: "codex", Skill: "a", TS: time.Unix(1, 0).UTC()}
	if err := s.Enqueue(ev); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	files, err := s.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("got %d files, want 1", len(files))
	}

	got, err := s.Read(files[0])
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Skill != "a" {
		t.Fatalf("read skill = %q", got.Skill)
	}

	if err := s.Remove(files[0]); err != nil {
		t.Fatalf("remove: %v", err)
	}
	files, _ = s.List()
	if len(files) != 0 {
		t.Fatalf("after remove got %d files, want 0", len(files))
	}
}

func TestSpoolListIgnoresTmpAndMarker(t *testing.T) {
	dir := t.TempDir()
	s := &Spool{Dir: dir}
	if err := os.WriteFile(filepath.Join(dir, "x.tmp"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, markerName), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	files, err := s.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("got %d files, want 0", len(files))
	}
}

func TestSpoolRotateDropsOldest(t *testing.T) {
	dir := t.TempDir()
	s := &Spool{Dir: dir}
	for i := 0; i < 5; i++ {
		if err := s.Enqueue(SkillEvent{Skill: "s", TS: time.Unix(int64(i), 0).UTC()}); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond) // ensure distinct nanos in filenames
	}
	dropped, err := s.Rotate(3)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if dropped != 2 {
		t.Fatalf("dropped = %d, want 2", dropped)
	}
	files, _ := s.List()
	if len(files) != 3 {
		t.Fatalf("remaining = %d, want 3", len(files))
	}
}

func TestSpoolRotateUnderCapNoop(t *testing.T) {
	dir := t.TempDir()
	s := &Spool{Dir: dir}
	_ = s.Enqueue(SkillEvent{Skill: "s", TS: time.Unix(1, 0).UTC()})
	dropped, err := s.Rotate(100)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if dropped != 0 {
		t.Fatalf("dropped = %d, want 0", dropped)
	}
}
