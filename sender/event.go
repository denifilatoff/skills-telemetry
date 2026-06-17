package main

import (
	"encoding/json"
	"time"
)

// SkillEvent is the normalized, agent-independent record produced by an adapter
// and persisted in the spool. It is the only shape that leaves this process.
type SkillEvent struct {
	Agent      string    `json:"agent"`
	SessionID  string    `json:"session_id"`
	RepoRemote string    `json:"repo_remote,omitempty"`
	Skill      string    `json:"skill"`
	TS         time.Time `json:"ts"`
}

func (e SkillEvent) MarshalJSON() ([]byte, error) {
	type alias SkillEvent
	return json.Marshal(alias(e))
}

func (e *SkillEvent) UnmarshalJSON(b []byte) error {
	type alias SkillEvent
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*e = SkillEvent(a)
	return nil
}
