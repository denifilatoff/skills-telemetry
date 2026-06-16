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

func TestDispatchUnknownAgent(t *testing.T) {
	if _, err := Dispatch("nope", []byte(`{}`), func(string) string { return "" }); err == nil {
		t.Fatal("want error for unknown agent")
	}
}
