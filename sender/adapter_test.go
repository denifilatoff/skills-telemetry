package main

import "testing"

func TestCodexAdapterParsesMarkers(t *testing.T) {
	stdin := []byte(`{
		"hook_event_name": "Stop",
		"session_id": "s1",
		"cwd": "/repo",
		"last_assistant_message": "done.\n[skill-called] skill=ops:deploy source=Netcracker/x\n[skill-called] skill=english-us-developer-style source=Netcracker/x\n"
	}`)
	events, err := Dispatch("codex", stdin, func(string) string { return "" })
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	if events[0].Skill != "ops:deploy" || events[0].Source != "Netcracker/x" {
		t.Fatalf("event[0] = %+v", events[0])
	}
	if events[0].Agent != "codex" || events[0].SessionID != "s1" {
		t.Fatalf("event[0] common fields = %+v", events[0])
	}
	if events[1].Skill != "english-us-developer-style" {
		t.Fatalf("event[1] = %+v", events[1])
	}
}

func TestCodexAdapterNoMarkers(t *testing.T) {
	events, err := Dispatch("codex", []byte(`{"last_assistant_message":"nothing here"}`), func(string) string { return "" })
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("got %d events, want 0", len(events))
	}
}

func TestCodexAdapterUsesRemoteResolver(t *testing.T) {
	stdin := []byte(`{"cwd":"/repo","last_assistant_message":"[skill-called] skill=a source=b"}`)
	events, err := Dispatch("codex", stdin, func(cwd string) string {
		if cwd != "/repo" {
			t.Fatalf("resolver got cwd %q", cwd)
		}
		return "git@host:org/repo.git"
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if events[0].RepoRemote != "git@host:org/repo.git" {
		t.Fatalf("remote = %q", events[0].RepoRemote)
	}
}

func TestClaudeAdapterParsesSkillTool(t *testing.T) {
	stdin := []byte(`{
		"session_id": "6a35f862",
		"transcript_path": "/Users/x/.claude/projects/p/6a35f862.jsonl",
		"cwd": "/repo",
		"hook_event_name": "PreToolUse",
		"tool_name": "Skill",
		"tool_input": {"skill": "superpowers:brainstorming", "args": "..."},
		"tool_use_id": "toolu_01"
	}`)
	events, err := Dispatch("claude", stdin, func(cwd string) string {
		if cwd != "/repo" {
			t.Fatalf("resolver got cwd %q", cwd)
		}
		return "git@host:org/repo.git"
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
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
	// Source is not recoverable from the native event; it stays empty.
	if e.Source != "" {
		t.Fatalf("source = %q, want empty", e.Source)
	}
}

func TestClaudeAdapterIgnoresOtherTools(t *testing.T) {
	stdin := []byte(`{"tool_name":"Bash","tool_input":{"command":"ls"}}`)
	events, err := Dispatch("claude", stdin, func(string) string { return "" })
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("got %d events, want 0", len(events))
	}
}

func TestClaudeAdapterMalformedJSON(t *testing.T) {
	events, err := Dispatch("claude", []byte(`{not json`), func(string) string { return "" })
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("got %d events, want 0", len(events))
	}
}

func TestDispatchUnknownAgent(t *testing.T) {
	if _, err := Dispatch("nope", []byte(`{}`), func(string) string { return "" }); err == nil {
		t.Fatal("want error for unknown agent")
	}
}

func TestCursorAdapterParsesMarker(t *testing.T) {
	stdin := []byte(`{
		"session_id": "c1",
		"text": "ok\n[skill-called] skill=ops:deploy source=Netcracker/x\n",
		"workspace_roots": ["/repo"],
		"transcript_path": "/nope.jsonl"
	}`)
	events, err := Dispatch("cursor", stdin, func(cwd string) string {
		if cwd != "/repo" {
			t.Fatalf("resolver got cwd %q", cwd)
		}
		return "git@host:org/repo.git"
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	e := events[0]
	if e.Agent != "cursor" || e.SessionID != "c1" || e.Skill != "ops:deploy" ||
		e.Source != "Netcracker/x" || e.RepoRemote != "git@host:org/repo.git" {
		t.Fatalf("event = %+v", e)
	}
}

func TestCursorAdapterNoMarker(t *testing.T) {
	events, err := Dispatch("cursor", []byte(`{"text":"nothing here"}`), func(string) string { return "" })
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("got %d events, want 0", len(events))
	}
}

func TestCursorAdapterMalformedJSON(t *testing.T) {
	events, err := Dispatch("cursor", []byte(`{not json`), func(string) string { return "" })
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("got %d events, want 0", len(events))
	}
}

func TestCursorAdapterParsesMultipleMarkers(t *testing.T) {
	stdin := []byte(`{
		"session_id": "c2",
		"text": "done.\n[skill-called] skill=ops:deploy source=Netcracker/x\n[skill-called] skill=english-developer-style source=Netcracker/x\n",
		"workspace_roots": ["/repo"]
	}`)
	events, err := Dispatch("cursor", stdin, func(string) string { return "" })
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	if events[0].Agent != "cursor" || events[0].Skill != "ops:deploy" {
		t.Fatalf("event[0] = %+v", events[0])
	}
	if events[1].Agent != "cursor" || events[1].Skill != "english-developer-style" {
		t.Fatalf("event[1] = %+v", events[1])
	}
}
