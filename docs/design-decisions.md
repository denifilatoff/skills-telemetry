# Design decisions

A short log of the main forks in this project and why each was taken. Each entry is the
decision, not the full analysis — the detailed records live under
[`docs/superpowers/`](superpowers/).

## How to detect a skill running

We compared eight ways to know that a skill ran in a session, from the agent's own
telemetry to rewriting skills as MCP calls. The full comparison is in
[the detection-options research](superpowers/research/2026-06-detection-options.md).

**Decision.** Use the agent's native skill event where one exists, and fall back to the
session transcript where it does not.

**Why.** No single method covers every agent. A native event is exact but only Claude
Code and OpenCode emit one; Codex and Cursor have nothing to intercept, so there the
skill is recognized from the skill-file reads in the transcript. See
[Agent integration](agent-integration.md) for the per-agent mechanics.

**The options we weighed.** Our detection is the combination of options 5 and 6: a
native event where the agent emits one, the transcript everywhere else.

| # | Approach | Verdict |
|---|---|---|
| 1 | The agent's own built-in telemetry | Only Claude Code names the skill; Codex, Cursor, and OpenCode have no skill breakdown. It is also indiscriminate — built-in telemetry is enabled globally for the agent, so it reports from every repository, not just the ones where telemetry was deliberately turned on. |
| 2 | A signal emitted from the skill body | Rejected — covert collection (it sends the moment the skill is read) and it pollutes a shared skill. |
| 3 | Routing a skill step through our MCP server | Rejected — hard-wires a portable skill to our server, and plain-text skills cannot be instrumented this way. |
| 4 | Subscribing to the agent's event stream | OpenCode only; the other three expose no interactive-session interface. |
| 5 | **A native skill-activation hook** ✅ | **Chosen where available** — exact, and the skill body stays untouched. Works on Claude Code (and OpenCode via `use_skill`); Codex and Cursor emit no such event. |
| 6 | **Reading the session transcript after the turn** ✅ | **Chosen as the fallback** for Codex and Cursor. Reliable on Claude Code and OpenCode, workable on Codex, the only route on Cursor. |
| 7 | A visible marker caught by a hook | Used at first, now retired — see [No marker in the skill output](#no-marker-in-the-skill-output). Probabilistic undercount, and it pollutes the model's output. |
| 8 | Langfuse as the backend instead of our stack | Rejected — it accepts only traces, with no skill concept except where activation is a tool call (OpenCode); Claude Code and Codex would need a central proxy that breaks subscription auth. |

## No marker in the skill output

**Decision.** Do not detect skills from a marker printed into the model's response.
Where there is no native event, rely on the session transcript instead.

**Why.** A marker makes the model emit a fixed line on every skill run, which
litters the visible output. It is worse when the skill is meant to return structured
data: an extra line corrupts the payload, and a reader who sees it may take the whole
response as malformed. The transcript records the same skill-file reads without
touching what the model prints, so we drop the marker entirely.

## Delivery and the consent boundary

**Decision.** Ship the hooks and the setup skill as two APM packages. The hooks
(`skills-telemetry` in
[`Netcracker/qubership-ai-packages`](https://github.com/Netcracker/qubership-ai-packages/tree/main/agent-packages/skills-telemetry))
are the package a repository depends on — installing it is the consent to send telemetry.
The setup skill (`skills-telemetry-configure` in this repository) is a dev dependency
for first-time provisioning on a new machine.

**Why.** The hooks are the same across every repository — three small JSON files, one per
harness. Hosting them in the shared Qubership marketplace lets any Qubership repository
add telemetry with a single `apm install` and no dependency on this CLI repository. The
setup skill and bootstrap scripts change with CLI releases, so they stay alongside the
CLI source. The hook calls the CLI by its bare name on `PATH`, so one hook command works
across every harness and OS.

## Provisioning the per-machine config

**Decision.** A `provision` subcommand in the Go binary does the security-sensitive
writes; a setup skill drives it in natural language. The full record is
[the provisioning decision](superpowers/decisions/2026-06-15-provisioning-and-paths.md).

**Why.** The endpoint, CA certificate, and token cannot ship in a public package, so
each machine sets them up once. The binary handles atomic writes, permissions, and
no-echo token entry; the skill handles discovery and interaction. The token is read by
the binary's no-echo prompt and never enters the model's context. A `curl | sh`
one-liner reuses the same binary for headless and CI contexts.

## Where config and buffered events live

**Decision.** Keep durable state (endpoint, token, CA, `machine.id`) in the config directory
and disposable state (the event outbox and transcript offsets) in the cache directory. The
binary itself is installed once on `PATH` at `~/.local/bin`, in neither. The full layout is in
[the skills-telemetry CLI](cli.md#file-layout).

**Why.** The cache holds only re-creatable state — the outbox re-fills on the next turn and the
offsets re-derive — so losing it under disk pressure is safe. Putting the token or endpoint
there would silently drop them and stop telemetry, so secrets and config live in the config
directory instead. Both directories resolve to **uniform XDG-style paths on every OS**
(`$XDG_CONFIG_HOME` else `~/.config`, `$XDG_CACHE_HOME` else `~/.cache`) rather than the
per-OS `os.UserConfigDir()` / `os.UserCacheDir()` locations. The original design used the
stdlib paths to avoid a custom resolver, but `os.UserConfigDir()` returns `%AppData%` on
Windows, which MSIX virtualizes for a packaged harness (Claude Desktop) — so a packaged and
a plain shell diverged onto different config dirs. A home-relative path outside `AppData` is
never virtualized, mirroring the binary's own `~/.local/bin`. The full record is
[the config-dir decision](superpowers/decisions/2026-06-23-config-cache-dir-xdg-msix.md).

## Event schema and privacy

**Decision.** Send the minimum: agent, session, repository remote, skill name, and an
anonymous install id. The full record is
[the event-schema decision](superpowers/decisions/2026-06-12-event-schema-and-privacy.md).

**Why.** The first end-to-end run leaked the local working directory
(`/Users/<username>/…`), so `repo.path` was dropped and a repository is now identified
by its remote URL alone. `machine.id` is a random UUID minted on first run, never
derived from the user or the hardware, so installs can be told apart without
identifying anyone. `turn.id` was dropped as finer than the session with no analytic
value.

## Transport security

**Decision.** TLS always, with no plaintext fallback. CA trust is additive.

**Why.** The event and the token must never leave the machine unencrypted. The
endpoint is always `https://`, certificate verification is never skipped, and a TLS
failure keeps the event in the outbox rather than downgrading. When a private CA is
provisioned the CLI appends it to the system trust pool, so a self-signed collector
works without replacing the system roots.

## Authentication

**Decision.** The CLI supports an optional bearer token. When a token is provisioned the
CLI sends it as `Authorization: Bearer`; without one the request carries no auth header.
A backend can verify the token to gate ingest.

**Why optional.** Not every deployment needs ingest auth, and the token has to reach each
machine out of band rather than through git. The client keeps it optional and the backend
decides whether to require it. How tokens are issued, distributed, and rotated is a
deployment concern, out of scope here.
