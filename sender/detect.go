package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// remoteResolver returns the git remote URL for a working dir, or "" if unknown.
// Injected so detectors stay pure and testable.
type remoteResolver func(cwd string) string

// detect routes a raw hook payload to the per-harness detector. Claude Code
// emits a native Skill-tool event; Codex and Cursor are detected from the
// session transcript.
func detect(agent string, stdin []byte, remote remoteResolver, now time.Time) ([]SkillEvent, error) {
	switch agent {
	case "claude":
		return claudeAdapter(stdin, remote, now)
	case "codex":
		return codexTranscriptEventsAuto(stdin, now), nil
	case "cursor":
		return cursorTranscriptEventsAuto(stdin, remote, now), nil
	default:
		return nil, fmt.Errorf("no detector for agent %q", agent)
	}
}

type codexPayload struct {
	SessionID      string `json:"session_id"`
	Cwd            string `json:"cwd"`
	TranscriptPath string `json:"transcript_path"`
}

// claudePayload is the Claude Code PreToolUse hook envelope. Only the fields
// the adapter needs are decoded; the rest (permission_mode, effort,
// tool_use_id, transcript_path) are ignored.
type claudePayload struct {
	SessionID string `json:"session_id"`
	Cwd       string `json:"cwd"`
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		Skill string `json:"skill"`
	} `json:"tool_input"`
}

// claudeAdapter turns a Claude Code PreToolUse hook payload into a single
// SkillEvent. The hook is registered on the Skill tool, so it fires once per
// skill activation; tool_input.skill is the skill name (namespace-prefixed for
// plugin skills, bare for project skills).
func claudeAdapter(stdin []byte, remote remoteResolver, now time.Time) ([]SkillEvent, error) {
	var p claudePayload
	if len(stdin) > 0 {
		// Malformed JSON yields no events rather than an error: a broken turn
		// must never fail the hook.
		_ = json.Unmarshal(stdin, &p)
	}
	// Defensive: the matcher already scopes the hook to the Skill tool, but
	// only emit when the payload confirms it and names a skill.
	if p.ToolName != "Skill" || p.ToolInput.Skill == "" {
		return nil, nil
	}
	// p.Cwd resolves the git remote only; the local path never leaves the
	// process, since it leaks the user's home directory and username.
	var rem string
	if remote != nil && p.Cwd != "" {
		rem = remote(p.Cwd)
	}
	return []SkillEvent{{
		Agent:      "claude",
		SessionID:  p.SessionID,
		RepoRemote: rem,
		Skill:      p.ToolInput.Skill,
		TS:         now,
	}}, nil
}

// cursorPayload is the Cursor afterAgentResponse hook envelope. Only the fields
// the adapter needs are decoded; the rest (conversation_id, generation_id,
// model, token counts, cursor_version, user_email) are ignored. user_email is
// deliberately not collected: it is PII, and the project drops repo.path and
// turn.id for the same reason.
type cursorPayload struct {
	SessionID      string   `json:"session_id"`
	WorkspaceRoots []string `json:"workspace_roots"`
	TranscriptPath string   `json:"transcript_path"`
}

// cursorRemote resolves the git remote from the first workspace root. Cursor
// gives no git data in the transcript, so the remote always comes from the hook
// payload.
func cursorRemote(p cursorPayload, remote remoteResolver) string {
	if remote == nil || len(p.WorkspaceRoots) == 0 || p.WorkspaceRoots[0] == "" {
		return ""
	}
	return remote(p.WorkspaceRoots[0])
}
