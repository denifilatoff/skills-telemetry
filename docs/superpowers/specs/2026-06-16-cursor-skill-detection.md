# Detecting skill use in Cursor sessions — 2026-06-16

How the sender recognizes that a skill ran inside a Cursor turn. Cursor adopted the Agent Skills
standard (`SKILL.md`) in version 2.4. There is no dedicated skill-activation event: a skill is
loaded either by reading its `SKILL.md` file (automatic activation) or by inlining its body into
the user message (manual `/skill-name`). The sender hangs detection off one low-frequency hook and
uses two complementary signals so that no single point of failure drops an event.

Scope: the Cursor harness only. Codex and Claude are detected differently and stay unchanged.

## Why this design over the alternatives

Three findings from a live probe on Cursor 3.6.31 (a throwaway hook set wired to every candidate
event, then a skill invoked both ways) settle the approach:

- **`afterAgentResponse` is the Cursor analog of the Codex `Stop` hook.** It fires once per agent
  response and carries everything the sender needs: the response `text`, the `transcript_path`, the
  `workspace_roots`, the `session_id`, and `cursor_version`. One hook, one ingest per response.
- **A skill load surfaces as a `Read` of its `SKILL.md`.** On automatic activation the agent reads
  `<root>/.cursor/skills/<name>/SKILL.md` through the `Read` tool, recorded in the transcript as an
  assistant `tool_use`. The skill name is the path segment. This is the same mechanism Codex uses,
  but Cursor hands over the transcript path directly, so the sender never has to locate the file.
- **The legacy SQLite store is not needed.** Cursor still keeps `state.vscdb`, but its schema is
  version-dependent and may not record skill activation. The JSONL transcript the hook points at is
  simple to parse and authoritative for the current session, so the sender ignores `state.vscdb`.

Rejected alternatives:

- **`preToolUse` path match.** Deterministic, but the hook fires on *every* file read, spawning the
  bootstrap on each one, and it misses manual `/skill-name` (no file read happens). The
  `afterAgentResponse` hook fires once per response and catches both activation modes.
- **Marker only.** Simplest, but probabilistic: it undercounts whenever a skill does not echo the
  marker. Kept as one of the two signals, not the only one.

## Two signals, kept together

Both signals run inside the single `afterAgentResponse` ingest and are deduplicated by skill name.

| Signal | Source | Catches | Misses |
|---|---|---|---|
| Marker | `[skill-called] skill=<name> source=<source>` echoed into the response, matched in the payload `text` | Turns where the model echoes the marker; the only signal that carries `source` | Turns where the model does not echo it |
| Transcript parse | The session JSONL at `transcript_path` | Every skill load, regardless of what the model writes | Nothing, within a turn the hook fires for |

## Where the transcript lives

The sender does not reconstruct the path. The `afterAgentResponse` payload carries it:

```
~/.cursor/projects/<project>/agent-transcripts/<conversation_id>/<conversation_id>.jsonl
```

The payload field is `transcript_path`. Read it directly.

## The parse rule

Each line is a JSON object. A skill load appears in one of two shapes, one per activation mode:

- **Automatic activation** — an assistant message whose content holds a `tool_use` with
  `name == "Read"` and `input.path` matching `(?:^|/)\.cursor/skills/([^/]+)/SKILL\.md`. The capture
  group is the skill name.
- **Manual `/skill-name`** — a user message whose text contains a `<manually_attached_skills>`
  block listing `Skill Name: <name>`. The body is inlined, so no `Read` occurs.

Both shapes yield a skill name and no `source`. The marker signal supplies `source` when present.

## When to parse, and deduplication

`afterAgentResponse` fires once per response and the transcript grows over the session, so the
sender keeps a per-session byte offset and parses only new lines. This reuses the existing
`OffsetStore` (the file is named `cursor_<session>.offset`). Findings from the marker and the
transcript are merged by skill name with `mergeBySkill`; the marker wins because it carries
`source`.

## Data-model mapping (`SkillEvent`)

| Field | Source |
|---|---|
| `agent` | `"cursor"`, from the hook registration `ingest --agent=cursor` |
| `session_id` | payload `session_id` (equals `conversation_id`) |
| `repo_remote` | `workspace_roots[0]` resolved through the existing `gitRemote` |
| `skill` | marker capture, or `SKILL.md` path segment, or `manually_attached_skills` name |
| `source` | marker only; empty otherwise |
| `ts` | send time |
| `machine.id`, `service.*` | sender constants, unchanged |

**`user_email` is not collected.** Cursor is the only harness that hands it over, but it is PII and
the project already drops `repo.path` and `turn.id` for the same reason. The anonymous `machine.id`
already tells installs apart. Capturing email would also break symmetry with the other harnesses.

`cursor_version` is available in the payload. The event schema has no agent-version field today, so
it is left out for now; adding one later is a separate change across all harnesses.

## What is reused versus new

Reused unchanged: the spool, flush, `gitRemote`, `machine.id`, `markerRe`, `mergeBySkill`, and
`OffsetStore`.

New: a `cursorAdapter` in `adapter.go` that decodes the `afterAgentResponse` payload and scans
`text` for the marker; a `transcript_cursor.go` parser for the two transcript shapes (mirroring
`transcript_codex.go`); registration of `"cursor"` in the `adapters` map; an `ingest` branch that
merges the marker events with the transcript events for `agent == "cursor"`; and the
`afterAgentResponse` hook filled into `skill-call-cursor-hooks.json`.

## Open questions

- **APM Cursor target.** Confirm that `apm install --target cursor` deploys
  `skill-call-cursor-hooks.json` to `<root>/.cursor/hooks.json` by the same naming convention used
  for Codex and Claude. The package has the hook file but no Cursor target has been exercised.
- **Cloud agents.** Cursor loads project hooks for cloud-agent runs too. Whether the bootstrap and
  provisioning model work in that environment is untested and out of scope for the first slice.

## Limitations

- The transcript schema is Cursor-internal. Pin to the fields above and treat an unknown shape as
  zero events, never as a hook failure.
- Manual `/skill-name` is detected only through the marker or the `manually_attached_skills` block;
  the block format is text, so it is the more fragile of the two transcript shapes.
