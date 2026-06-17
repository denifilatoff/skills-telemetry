# CLAUDE.md

**Read [README.md](README.md) first**, then this file. The README is the entry point — what
the project does, how telemetry is turned on, the architecture, the data, and the backend
requirements. Deeper docs are linked from its Documentation section. This file holds only
what an agent needs beyond the README: orientation, conventions, and open work.

## Orientation

Skill-usage telemetry for AI coding agents. A skill run is detected per harness, sent to a
shared OpenTelemetry collector, and packaged through APM so that installing the package into
a repository is the consent boundary.

- **Component:** the `skills-telemetry` CLI — a small Go binary in `sender/`. It detects the
  skill, buffers events to a local spool, and flushes over OTLP/HTTPS. No daemon. See
  [docs/cli.md](docs/cli.md).
- **Detection:** a native hook event where the agent emits one (Claude Code), the session
  transcript otherwise (Codex, Cursor). See [docs/agent-integration.md](docs/agent-integration.md).
- **Harnesses:** Codex, Claude Code, and Cursor are shipped (v0.5.0). OpenCode is planned.
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
  `.codex/`, `.claude/`, `.cursor/`, `apm.lock.yaml`, `sender/dist/`,
  `sender/skills-telemetry`, and `eval-workspace/`.
- **Keep:** the root `apm.yml` — gitignored and machine-specific, but the install needs it —
  and the per-machine config outside the repo (endpoint, CA, token, `machine.id` under the
  config dir).

Do not run `git clean -xdf` blindly: it would also delete the root `apm.yml` and any
untracked files not yet committed. Remove the listed paths explicitly.

## Open work

- **Drop the `[skill-called]` marker from the code.** The docs already treat it as retired,
  but the CLI still implements it: `markerRe` and the marker adapters in `sender/adapter.go`,
  the marker dedup in `sender/transcript_codex.go` and `sender/main.go`, plus the tests and
  the `agent-packages/adr-authoring` fixture that emits it.
- **Rename "sender" to the skills-telemetry CLI in the code.** The source directory `sender/`
  (and its paths in `.github/workflows/release.yml`), the Go comments, and the emitted
  `service.name` (`qubership-skills-telemetry-sender`) still use the old name. Renaming
  `service.name` is backend-visible — dashboards key on it — so decide it deliberately. The Go
  module is already `skills-telemetry`.
- **OpenCode adapter** — the fourth harness. A native `use_skill` tool call via the
  `.claude/skills/` compatibility extension, the same path as Claude Code.
- **Token auth** — the collector is unauthenticated. The CLI already sends a bearer token; the
  gateway must verify it. The shared-versus-per-user-token fork is still open (see
  [Authentication in the design decisions](docs/design-decisions.md#authentication-open)).
- **Spool housekeeping** — offset-file garbage collection is not implemented.
