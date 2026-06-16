package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

// remoteResolver returns the git remote URL for a working dir, or "" if unknown.
// Injected so adapters stay pure and testable.
type remoteResolver func(cwd string) string

// Adapter turns a harness-specific hook payload into normalized events.
type Adapter func(stdin []byte, remote remoteResolver, now time.Time) ([]SkillEvent, error)

var adapters = map[string]Adapter{
	"codex": codexAdapter,
}

// Dispatch routes raw stdin to the adapter for the named agent.
func Dispatch(agent string, stdin []byte, remote remoteResolver) ([]SkillEvent, error) {
	a, ok := adapters[agent]
	if !ok {
		return nil, fmt.Errorf("no adapter for agent %q", agent)
	}
	return a(stdin, remote, time.Now().UTC())
}

var markerRe = regexp.MustCompile(`(?m)^\[skill-called\]\s+skill=(\S+)\s+source=(\S+)\s*$`)

type codexPayload struct {
	SessionID            string `json:"session_id"`
	Cwd                  string `json:"cwd"`
	LastAssistantMessage string `json:"last_assistant_message"`
	// TranscriptPath is the rollout file Codex passes to the Stop hook. The
	// transcript adapter parses it for SKILL.md reads; the marker adapter
	// ignores it. No glob by session id is needed.
	TranscriptPath string `json:"transcript_path"`
}

func codexAdapter(stdin []byte, remote remoteResolver, now time.Time) ([]SkillEvent, error) {
	var p codexPayload
	if len(stdin) > 0 {
		// Malformed JSON yields no events rather than an error: a broken turn
		// must never fail the hook.
		_ = json.Unmarshal(stdin, &p)
	}
	matches := markerRe.FindAllStringSubmatch(p.LastAssistantMessage, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	// p.Cwd is used only to resolve the git remote; the local path itself is
	// never emitted, since it leaks the user's home directory and username.
	var rem string
	if remote != nil && p.Cwd != "" {
		rem = remote(p.Cwd)
	}
	events := make([]SkillEvent, 0, len(matches))
	for _, m := range matches {
		events = append(events, SkillEvent{
			Agent:      "codex",
			SessionID:  p.SessionID,
			RepoRemote: rem,
			Skill:      m[1],
			Source:     m[2],
			TS:         now,
		})
	}
	return events, nil
}
