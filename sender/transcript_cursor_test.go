package main

import (
	"strings"
	"testing"
)

func TestScanCursorTranscriptReadToolUse(t *testing.T) {
	line := `{"role":"assistant","message":{"content":[` +
		`{"type":"text","text":"reading skill"},` +
		`{"type":"tool_use","name":"Read","input":{"path":"/repo/.cursor/skills/adr-authoring/SKILL.md"}}` +
		`]}}` + "\n"
	skills, end := scanCursorTranscript(strings.NewReader(line), 0)
	if len(skills) != 1 || skills[0] != "adr-authoring" {
		t.Fatalf("skills = %v", skills)
	}
	if end != int64(len(line)) {
		t.Fatalf("end = %d, want %d", end, len(line))
	}
}

func TestScanCursorTranscriptManualAttach(t *testing.T) {
	line := `{"role":"user","message":{"content":[{"type":"text","text":` +
		`"<manually_attached_skills>\nSkill Name: telemetry-probe\nPath: /x/SKILL.md\n"}]}}` + "\n"
	skills, _ := scanCursorTranscript(strings.NewReader(line), 0)
	if len(skills) != 1 || skills[0] != "telemetry-probe" {
		t.Fatalf("skills = %v", skills)
	}
}

func TestScanCursorTranscriptIgnoresNonSkillReadsAndDedups(t *testing.T) {
	lines := `{"role":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"path":"/repo/src/main.go"}}]}}` + "\n" +
		`{"role":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"path":"/repo/.cursor/skills/a/SKILL.md"}}]}}` + "\n" +
		`{"role":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"path":"/repo/.cursor/skills/a/SKILL.md"}}]}}` + "\n"
	skills, _ := scanCursorTranscript(strings.NewReader(lines), 0)
	if len(skills) != 1 || skills[0] != "a" {
		t.Fatalf("skills = %v", skills)
	}
}

func TestScanCursorTranscriptOffsetSkipsEarlyLines(t *testing.T) {
	first := `{"role":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"path":"/repo/.cursor/skills/old/SKILL.md"}}]}}` + "\n"
	second := `{"role":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"path":"/repo/.cursor/skills/new/SKILL.md"}}]}}` + "\n"
	skills, _ := scanCursorTranscript(strings.NewReader(first+second), int64(len(first)))
	if len(skills) != 1 || skills[0] != "new" {
		t.Fatalf("skills = %v", skills)
	}
}

func TestScanCursorTranscriptSkipsMalformedLine(t *testing.T) {
	lines := "{not json\n" +
		`{"role":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"path":"/repo/.cursor/skills/a/SKILL.md"}}]}}` + "\n"
	skills, _ := scanCursorTranscript(strings.NewReader(lines), 0)
	if len(skills) != 1 || skills[0] != "a" {
		t.Fatalf("skills = %v", skills)
	}
}

func TestScanCursorTranscriptManualAttachMultiple(t *testing.T) {
	line := `{"role":"user","message":{"content":[{"type":"text","text":` +
		`"<manually_attached_skills>\nSkill Name: alpha\nPath: /a\nSkill Name: beta\nPath: /b\n"}]}}` + "\n"
	skills, _ := scanCursorTranscript(strings.NewReader(line), 0)
	if len(skills) != 2 || skills[0] != "alpha" || skills[1] != "beta" {
		t.Fatalf("skills = %v", skills)
	}
}
