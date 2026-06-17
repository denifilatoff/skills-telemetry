package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDetectClaudeParsesSkillTool(t *testing.T) {
	stdin := []byte(`{"session_id":"6a35f862","cwd":"/repo","tool_name":"Skill","tool_input":{"skill":"superpowers:brainstorming"}}`)
	events, err := detect("claude", stdin, func(cwd string) string {
		if cwd != "/repo" {
			t.Fatalf("resolver got cwd %q", cwd)
		}
		return "git@host:org/repo.git"
	}, time.Now().UTC())
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	e := events[0]
	if e.Agent != "claude" || e.SessionID != "6a35f862" || e.Skill != "superpowers:brainstorming" {
		t.Fatalf("event = %+v", e)
	}
	if e.RepoRemote != "git@host:org/repo.git" {
		t.Fatalf("remote = %q", e.RepoRemote)
	}
}

func TestDetectClaudeIgnoresOtherTools(t *testing.T) {
	events, err := detect("claude", []byte(`{"tool_name":"Bash","tool_input":{"command":"ls"}}`), func(string) string { return "" }, time.Now().UTC())
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("got %d events, want 0", len(events))
	}
}

func TestDetectClaudeMalformedJSON(t *testing.T) {
	events, err := detect("claude", []byte(`{not json`), func(string) string { return "" }, time.Now().UTC())
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("got %d events, want 0", len(events))
	}
}

func TestDetectUnknownAgent(t *testing.T) {
	if _, err := detect("nope", []byte(`{}`), func(string) string { return "" }, time.Now().UTC()); err == nil {
		t.Fatal("want error for unknown agent")
	}
}

func TestDetectCodexFromTranscript(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	tp := filepath.Join(t.TempDir(), "r.jsonl")
	body := `{"type":"session_meta","payload":{"git":{"repository_url":"git@host:o/r.git"}}}` + "\n" +
		`{"type":"response_item","payload":{"type":"function_call","name":"exec_command","arguments":"{\"cmd\":\"cat skills/demo/SKILL.md\"}"}}` + "\n"
	if err := os.WriteFile(tp, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	stdin := []byte(`{"session_id":"s1","transcript_path":"` + tp + `"}`)
	events, err := detect("codex", stdin, func(string) string { return "" }, time.Now().UTC())
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if len(events) != 1 || events[0].Agent != "codex" || events[0].Skill != "demo" {
		t.Fatalf("events = %+v", events)
	}
	if events[0].RepoRemote != "git@host:o/r.git" {
		t.Fatalf("remote = %q", events[0].RepoRemote)
	}
}

func TestDetectCursorFromTranscript(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	tp := filepath.Join(t.TempDir(), "t.jsonl")
	body := `{"message":{"content":[{"type":"tool_use","name":"Read","input":{"path":"/repo/.cursor/skills/demo/SKILL.md"}}]}}` + "\n"
	if err := os.WriteFile(tp, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	stdin := []byte(`{"session_id":"c1","workspace_roots":["/repo"],"transcript_path":"` + tp + `"}`)
	events, err := detect("cursor", stdin, func(string) string { return "git@host:o/r.git" }, time.Now().UTC())
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if len(events) != 1 || events[0].Agent != "cursor" || events[0].Skill != "demo" {
		t.Fatalf("events = %+v", events)
	}
	if events[0].RepoRemote != "git@host:o/r.git" {
		t.Fatalf("remote = %q", events[0].RepoRemote)
	}
}
