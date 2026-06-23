# CLAUDE.md

**Read [README.md](README.md) first**, then this file. The README is the entry point — what
the project does, how telemetry is turned on, the architecture, the data, and the backend
requirements. Deeper docs are linked from its Documentation section. This file holds only
what an agent needs beyond the README: orientation, conventions, and open work.

## Orientation

Skill-usage telemetry for AI coding agents. A skill run is detected per harness, sent to a
shared OpenTelemetry collector, and packaged through APM so that installing the package into
a repository is the consent boundary.

- **Component:** the `skills-telemetry` CLI — a small Go binary at the repository root
  (a flat `package main`, the "Basic command" layout from the Go module-layout guide). It
  detects the skill, buffers events to a local outbox, and flushes over OTLP/HTTPS. No
  daemon. See [docs/cli.md](docs/cli.md).
- **Invocation: the binary is on `PATH`, not behind a bootstrap wrapper.** The harness and
  the `provision-skills-telemetry` skill call the binary directly — on a dev machine it lives
  at `~/.local/bin/skills-telemetry` (current build reports `0.6.0-dev`). The old
  `bootstrap.{sh,ps1,bat}` wrapper that located-or-downloaded a version-pinned release is
  **retired**: it still pins an older tag (e.g. `v0.5.3`) and downloads into `LOCALAPPDATA`,
  so running `status` through it reports a stale binary and a misleading version. To check
  state, run `skills-telemetry status` / `version` against the on-`PATH` binary, never the
  bootstrap script. **This is in a testing phase:** the direct on-`PATH` invocation is being
  trialled across all harnesses (Codex, Claude Code, Cursor) before it replaces the bootstrap
  wrapper everywhere — expect both paths to coexist until the rollout is confirmed.
- **Detection:** a native hook event where the agent emits one (Claude Code), the session
  transcript otherwise (Codex, Cursor). See [docs/agent-integration.md](docs/agent-integration.md).
- **Harnesses:** Codex, Claude Code, and Cursor are shipped (v0.5.0). OpenCode is planned.
- **Config & cache paths: uniform XDG, not `os.UserConfigDir()`.** Durable config lives at
  `$XDG_CONFIG_HOME` else `~/.config/qubership-skills-telemetry/` and the spool at
  `$XDG_CACHE_HOME` else `~/.cache/qubership-skills-telemetry/` — the same path on every OS,
  mirroring the binary's `~/.local/bin`. This is deliberate: `os.UserConfigDir()` is
  `%AppData%` on Windows, which MSIX virtualizes for a packaged harness (Claude Desktop), so a
  packaged and a plain shell diverged onto different config dirs. A home-relative path outside
  `AppData` is never virtualized. Resolved in [config.go](config.go) (`configBase`) and
  [outbox.go](outbox.go) (`cacheBase`); rationale in
  [docs/superpowers/decisions/2026-06-23-config-cache-dir-xdg-msix.md](docs/superpowers/decisions/2026-06-23-config-cache-dir-xdg-msix.md).
  The binary does **not** auto-migrate; the `provision-skills-telemetry` skill documents moving
  an existing AppData/Library install to the new location.
- **Out of scope:** the collector, gateway, and storage (VictoriaMetrics, VictoriaLogs,
  Grafana) are infrastructure.
- **Decisions:** the main forks and why each was taken are in
  [docs/design-decisions.md](docs/design-decisions.md); historical records sit under
  `docs/superpowers/`.

## Conventions

- **English only.** Every committed file — Markdown, code, comments, commit messages,
  identifiers — is English. Translate anything else before committing.
- **Docs vs history.** Current, maintained documentation lives in `docs/` and the README.
  `docs/superpowers/` is a working archive — dated specs, plans, decisions, and research
  that are snapshots and are not kept up to date. When something changes, update `docs/`;
  never edit a dated `docs/superpowers/` file to match.
- **Naming.** The component is the "skills-telemetry CLI". The response-text "marker" is
  retired terminology — never reintroduce "breadcrumb".
- **Present design forks via AskUserQuestion**, recommendation first, and expect the
  recommendation to be challenged.
- **APM gotchas.** Install with `apm install --target <codex|claude|cursor|all>`, then
  `apm compile`. Cursor needs `.cursor/` to exist before install, and a fresh `apm install`
  drops the required top-level `version` from `.cursor/hooks.json` — re-add `"version": 1`
  until [microsoft/apm#1823](https://github.com/microsoft/apm/issues/1823) ships. APM-generated
  artifacts (`apm_modules/`, `.agents/`, `.codex/`, `.claude/`, `.cursor/`, `apm.lock.yaml`)
  are gitignored; do not commit them.

## Git workflow

Solo repository, so the path scales to the change:

- **Minor** (docs, `.gitignore`, small fixes — nothing users see): `commit` → `push`
  straight to `main`. No branch, no PR.
- **Significant** (features, user-visible changes): branch → `commit` → `push` → PR
  (squash) → auto-merge once CI is green → `tag`. The PR buys the CI gate and a
  transparent, revertible history; the tag triggers the release.

Keep history linear (squash merges) and commit messages in Conventional Commits. `main`
has no strict branch protection on purpose: the release workflow pushes version bumps
back to it, which required reviews would block.

## Testing and cleanup

A test run is `apm install` / `apm compile`, exercising the hook, then removing the
generated files so the next run starts clean. They are all gitignored; preview them with
`git clean -xdn`.

- **Remove** (APM install artifacts and build output): `apm_modules/`, `.agents/`,
  `.codex/`, `.claude/`, `.cursor/`, `apm.lock.yaml`, `dist/`, the root
  `skills-telemetry` binary, and `eval-workspace/`.
- **Keep:** the root `apm.yml` — gitignored and machine-specific, but the install needs it —
  and the per-machine config outside the repo (endpoint, CA, token, `machine.id` under the
  config dir).

Do not run `git clean -xdf` blindly: it would also delete the root `apm.yml` and any
untracked files not yet committed. Remove the listed paths explicitly.

## Open work

- **OpenCode adapter** — the fourth harness. A native `use_skill` tool call via the
  `.claude/skills/` compatibility extension, the same path as Claude Code.
- **Outbox housekeeping** — offset-file garbage collection is not implemented.
- **Dashboards.** The OTLP `service.name` changed from `qubership-skills-telemetry-sender`
  to `qubership-skills-telemetry`; update the Grafana key that still references the old value.
