package main

import (
	"bufio"
	"encoding/json"
	"io"
	"regexp"
	"strings"
)

// cursorSkillReadRe matches the Read tool input.path of a skill body. The path
// is absolute, ending in .cursor/skills/<name>/SKILL.md. The character before
// `.cursor` must be a separator (or start of string) so a path like
// `my.cursor/...` does not match.
var cursorSkillReadRe = regexp.MustCompile(`(?:^|/)\.cursor/skills/([^/]+)/SKILL\.md`)

// cursorManualSkillRe matches a `Skill Name: <name>` line inside the
// <manually_attached_skills> block Cursor inlines on a manual /skill-name call.
var cursorManualSkillRe = regexp.MustCompile(`(?m)^Skill Name:\s*(\S+)`)

// scanCursorTranscript streams a Cursor transcript and returns the skill names
// read at or beyond startOffset, in order and deduped, plus the end-of-file byte
// offset to persist as the next offset. Unlike Codex, the repo remote is not in
// the transcript, so it is resolved from the hook payload instead.
func scanCursorTranscript(r io.Reader, startOffset int64) ([]string, int64) {
	var skills []string
	seen := map[string]bool{}
	br := bufio.NewReader(r)
	var pos int64
	for {
		line, err := br.ReadString('\n')
		if len(line) > 0 {
			lineStart := pos
			pos += int64(len(line))
			if lineStart >= startOffset {
				processCursorLine(line, &skills, seen)
			}
		}
		if err != nil {
			break
		}
	}
	return skills, pos
}

func processCursorLine(line string, skills *[]string, seen map[string]bool) {
	var env struct {
		Message struct {
			Content []struct {
				Type  string `json:"type"`
				Name  string `json:"name"`
				Text  string `json:"text"`
				Input struct {
					Path string `json:"path"`
				} `json:"input"`
			} `json:"content"`
		} `json:"message"`
	}
	if json.Unmarshal([]byte(line), &env) != nil {
		return
	}
	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			*skills = append(*skills, name)
		}
	}
	for _, c := range env.Message.Content {
		switch c.Type {
		case "tool_use":
			if c.Name != "Read" {
				continue
			}
			if m := cursorSkillReadRe.FindStringSubmatch(c.Input.Path); m != nil {
				add(m[1])
			}
		case "text":
			if !strings.Contains(c.Text, "<manually_attached_skills>") {
				continue
			}
			for _, m := range cursorManualSkillRe.FindAllStringSubmatch(c.Text, -1) {
				add(m[1])
			}
		}
	}
}
