# The skills-telemetry CLI

The `skills-telemetry` CLI is the local component that detects skill use, buffers
events to a local outbox, and forwards them to the collector over OTLP/HTTPS. It runs
from the agent hook on each turn — there is no daemon and no background process.

To install it, see [Installation in the README](../README.md#installation). This document
covers the command reference, how it works inside, and where it keeps its files.

## What it does

On each hook run the CLI:

1. reads the agent's hook payload from stdin and detects any skill that ran (see
   [Agent integration](agent-integration.md));
2. normalizes the detection into one OpenTelemetry log record per skill run (see
   [Data](../README.md#data));
3. writes each record to the on-disk outbox;
4. opportunistically flushes the outbox to the collector over OTLP/HTTPS.

## Subcommands

The hook calls `ingest`; the setup skill calls the rest, so you rarely run them by hand.
`skills-telemetry <command>`:

| Command | Purpose |
|---|---|
| `provision` | Write the per-machine config: collector endpoint, optional CA certificate (`--ca=<path>`), and an optional token read without echo. Idempotent. |
| `status` | Read-only check: build version, config directory, endpoint, whether a CA is present, outbox backlog, last flush attempt, and a provisioned verdict. Sends nothing. |
| `selftest` | Send one marked probe event and report whether the collector accepted it and it left the outbox. |
| `ingest` | The hook path: read an agent hook payload on stdin, detect skill use (on Codex the `SKILL.md` reads in the session rollout; on Claude Code the `Skill` tool name in the `PreToolUse` payload; on Cursor the `SKILL.md` reads in the `afterAgentResponse` transcript), queue the events, and flush opportunistically. Always exits 0 so it never fails an agent turn. |
| `flush` | Send queued events to the collector and delete each on success. |
| `version` | Print the build version. |

## Buffering and delivery

The CLI never blocks an agent turn on the network. `ingest` writes events to the outbox
and returns; delivery happens opportunistically.

- **Outbox.** One JSON file per buffered event in a per-machine outbox directory. A failed
  send leaves the file in place to retry on a later run.
- **Flush lock.** `flush` takes a non-blocking advisory lock (`.flush.lock`) on the
  outbox, so two concurrent runs never send the same event twice. A run that finds the
  lock held skips quietly.
- **Offset dedup.** On harnesses detected from the transcript (Codex, Cursor), the CLI
  stores a per-session byte offset into the transcript and reads only the lines written
  since the previous run. The key is namespaced per harness (`codex:<session>`), so
  different harnesses do not collide and one skill run counts once.

## Transport and security

Delivery is OTLP/HTTP, over HTTPS only. The CLI never falls back to plaintext and never
skips certificate verification; a TLS failure keeps the event in the outbox rather than
downgrading. When a private CA is provisioned, the CLI appends it to the system trust
pool — trust stays additive — so a self-signed collector works without replacing the
system roots. A per-machine access token is optional: when provisioned, the CLI sends it
as an `Authorization: Bearer` header; without one, the request carries no auth header.

## File layout

The CLI splits durable state from disposable state, following the platform conventions
(`os.UserConfigDir()` and `os.UserCacheDir()`). The reasoning is in
[the provisioning decision](superpowers/decisions/2026-06-15-provisioning-and-paths.md).

| Location | Path | Holds |
|---|---|---|
| **Config** (durable) | `<UserConfigDir>/qubership-skills-telemetry/` | `env` (endpoint, token), `ca.crt` (optional private CA), `machine-id` (anonymous install UUID) |
| **Cache** (disposable) | `<UserCacheDir>/qubership-skills-telemetry/` | the binary, `outbox/` (one JSON file per event, plus `.lastflush` and `.flush.lock`), `offsets/` (per-session transcript offsets) |

`<UserConfigDir>` is `~/Library/Application Support` on macOS, `~/.config` on Linux, and
`%AppData%` on Windows. Config holds anything that must survive — losing it stops
telemetry — so the token and endpoint never live in the cache, which the OS may purge
under disk pressure.
