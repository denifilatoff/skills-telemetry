package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestIngestEnqueuesAndFlushes(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Isolate the config dir so a CA provisioned on the dev machine does not
	// force TLS onto the plain-HTTP test server (caTLSConfig reads pkgConfigDir).
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	s := &Spool{Dir: t.TempDir()}
	stdin := []byte(`{"session_id":"s1","cwd":"/repo","last_assistant_message":"[skill-called] skill=a source=b"}`)

	code := ingest(s, "codex", srv.URL, stdin, func(string) string { return "" })
	if code != 0 {
		t.Fatalf("ingest exit = %d, want 0", code)
	}
	if atomic.LoadInt32(&hits) == 0 {
		t.Fatal("expected a flush on first ingest")
	}
	files, _ := s.List()
	if len(files) != 0 {
		t.Fatalf("buffer should be drained: %d files", len(files))
	}
	if _, err := os.Stat(filepath.Join(s.Dir, markerName)); err != nil {
		t.Fatalf("throttle marker missing: %v", err)
	}
}

func TestIngestBadJSONStillSucceeds(t *testing.T) {
	s := &Spool{Dir: t.TempDir()}
	code := ingest(s, "codex", "", []byte("not json"), func(string) string { return "" })
	if code != 0 {
		t.Fatalf("ingest exit = %d, want 0", code)
	}
}

func TestIngestCursorMergesMarkerAndTranscript(t *testing.T) {
	// Isolate config/cache dirs so the real machine state and any provisioned CA
	// are untouched (DefaultOffsetStore uses the cache dir; caTLSConfig the config dir).
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	dir := t.TempDir()
	tp := filepath.Join(dir, "t.jsonl")
	// Transcript reads skill "from-transcript"; the marker reports "from-marker".
	body := `{"role":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"path":"/repo/.cursor/skills/from-transcript/SKILL.md"}}]}}` + "\n"
	if err := os.WriteFile(tp, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	stdin := []byte(`{"session_id":"c1","workspace_roots":["/repo"],` +
		`"text":"[skill-called] skill=from-marker source=Netcracker/x",` +
		`"transcript_path":"` + tp + `"}`)

	// Empty endpoint => Flush is a no-op, so events stay in the spool to inspect.
	s := &Spool{Dir: t.TempDir()}
	if code := ingest(s, "cursor", "", stdin, func(string) string { return "" }); code != 0 {
		t.Fatalf("ingest exit = %d, want 0", code)
	}

	files, err := s.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	// Two distinct skills: from-marker (signal 1) and from-transcript (signal 2).
	if len(files) != 2 {
		t.Fatalf("spooled %d events, want 2", len(files))
	}
}

func TestShouldFlushThrottle(t *testing.T) {
	dir := t.TempDir()
	s := &Spool{Dir: dir}
	if shouldFlush(s, 10, time.Hour) {
		t.Fatal("should not flush with empty buffer")
	}
	_ = s.Enqueue(SkillEvent{Skill: "x", TS: time.Now().UTC()})
	if !shouldFlush(s, 10, time.Hour) {
		t.Fatal("should flush when no prior attempt")
	}
	touchMarker(s)
	if shouldFlush(s, 10, time.Hour) {
		t.Fatal("should skip: marker fresh and count below N")
	}
}
