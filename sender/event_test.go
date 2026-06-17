package main

import (
	"testing"
	"time"
)

func TestSkillEventJSONRoundTrip(t *testing.T) {
	ts := time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC)
	in := SkillEvent{
		Agent:      "codex",
		SessionID:  "s1",
		RepoRemote: "git@host:org/repo.git",
		Skill:      "ops:deploy",
		TS:         ts,
	}
	b, err := in.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out SkillEvent
	if err := out.UnmarshalJSON(b); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out != in {
		t.Fatalf("round trip mismatch:\n got %+v\nwant %+v", out, in)
	}
}
