# CLAUDE.md

## Where this is going

This repository builds skill-usage telemetry for AI coding agents. The path:

1. **Detect** skill execution per harness — done for Codex via a `Stop` hook that reads the
   `[skill-called] skill=<name> source=<source>` marker.
2. **Send** those events to a shared collector — a small Go CLI (the "sender") that
   normalizes each event, buffers it in a machine-global spool, and flushes over OTLP/HTTP.
   See `docs/superpowers/specs/2026-06-12-local-telemetry-sender-design.md` and the plan in
   `docs/superpowers/plans/`.
3. **Package** the hook and sender through APM so installing the package into a repository is
   the consent boundary for sending telemetry.

The collector, gateway, and storage (VictoriaMetrics, VictoriaLogs, Grafana) are
infrastructure and are out of scope for now. Current focus: Codex; Claude, Cursor, and
OpenCode adapters are follow-up work.

## Language policy

- **All repository files are English only.** Markdown, code, comments, commit messages, and
  identifiers — no other language in any committed file. Translate anything that is not
  English before committing.
