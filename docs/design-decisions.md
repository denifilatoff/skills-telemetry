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

**Decision.** Ship the hook and the `skills-telemetry` CLI as one APM package, installed
per repository. Installing it is the consent to send telemetry.

**Why.** APM is project-scoped, so installing into a work repository keeps the scope
narrow and separates work from non-work usage. The CLI rides along as a package
resource and the hook calls it by a path relative to the project root, so there is no
machine-global install to manage. The collector address is fixed in the hook text, so
where the data goes is visible and not configurable.

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

**Decision.** Keep durable state (endpoint, token, CA, `machine.id`) in the OS config
directory, and disposable state (the binary and the event outbox) in the cache directory.
The full layout is in [the skills-telemetry CLI](cli.md#file-layout).

**Why.** The cache is disposable by design — the binary re-downloads, and macOS purges
`~/Library/Caches` under disk pressure. Putting the token or endpoint there would
silently drop them and stop telemetry, so secrets and config live in the config
directory instead. Both directories follow the platform conventions
(`os.UserConfigDir()`, `os.UserCacheDir()`), so the CLI needs no custom path resolver.

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

**Why.** The event and the future token must never leave the machine unencrypted. The
endpoint is always `https://`, certificate verification is never skipped, and a TLS
failure keeps the event in the outbox rather than downgrading. When a private CA is
provisioned the CLI appends it to the system trust pool, so a self-signed collector
works without replacing the system roots.

## Authentication (open)

**Decision.** Not settled. The CLI already sends `Authorization: Bearer` from a
per-machine token, and provisioning accepts one, but the gateway does not yet verify it.

**Why it is open.** The token has to reach each machine without going through git, which
is in tension with having no central machine management. The unresolved fork is a shared
write-only token for everyone versus a personal token per participant — personal is
better for attribution and revocation, shared is simpler to distribute. Where a
participant obtains the token (most likely self-service through an internal portal behind
corporate sign-on) and how it is rotated and revoked are part of the same fork. To be
settled before rollout.
