# qubership-skills-telemetry

This package provides target-specific hook implementations for observing
visible skill invocation markers.

Current status:

- Codex: implemented fully
- Cursor: placeholder
- Claude: placeholder

The Codex `Stop` hook calls `bootstrap.sh` (macOS/Linux) or `bootstrap.ps1`
(Windows). The bootstrap fetches the pinned `skills-telemetry` Go binary into
a per-machine cache on first run, then runs `ingest --agent=codex`. The
`ingest` command reads the hook payload from stdin, normalizes the event, and
writes it to a machine-global spool. The same run opportunistically flushes
buffered events to the collector over OTLP/HTTP (no daemon).

Codex is implemented. Claude, Cursor, and OpenCode adapters are follow-up work.

## Configuration

The sender reads its collector settings from the environment, delivered per
machine out of band (a secret manager or onboarding step, never git):

- `SKILLS_TELEMETRY_ENDPOINT` — the OTLP/HTTP collector URL, for example
  `https://collector.example/v1/logs`. Without it the flush is a no-op, so
  events stay buffered in the spool.
- `SKILLS_TELEMETRY_TOKEN` — the bearer token sent as `Authorization: Bearer`.
  Falls back to a secret file at
  `<user-config-dir>/qubership-skills-telemetry/token`. Without it the request
  carries no auth header.

An explicit `--endpoint=<url>` flag on the hook command overrides
`SKILLS_TELEMETRY_ENDPOINT` when you want to pin the collector for one
repository.

## Release

Binaries are built and published by the `release` GitHub Actions workflow, not
on a local machine. Push a `v*` tag to `denifilatoff/skills-telemetry`:

```
git tag v0.1.0 && git push origin v0.1.0
```

The workflow runs the sender tests, cross-compiles six targets (darwin, linux,
and windows, each for amd64 and arm64), writes `SHA256SUMS`, and attaches every
artifact to a GitHub Release. `bootstrap.sh` and `bootstrap.ps1` download
`skills-telemetry-<os>-<arch>` from that release; keep `BINARY_VERSION` in both
scripts in step with the tag you push.

Checksum verification in the bootstrap scripts is follow-up work; the release
already publishes `SHA256SUMS` for it.
