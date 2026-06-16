# adr-authoring

An APM package that helps agents write Architecture Decision Records following the
Qubership ADR contract: one decision per record, stored under `docs/adr/` with a
`NNNN-kebab-slug.md` filename, a `Proposed | Accepted | Rejected | Superseded`
status, ISO 8601 dates, and immutable accepted records superseded by new ones.

## Contents

- `.apm/instructions/adr-authoring.instructions.md` — the trigger merged into
  `AGENTS.md` / `CLAUDE.md` by `apm compile`.
- `.apm/skills/adr-authoring/SKILL.md` — the on-demand how-to: local conventions
  plus the ADR template.

## Note

This package doubles as the instrumented test skill for `qubership-skills-telemetry`.
Its `SKILL.md` emits a `[skill-called]` marker on activation, which the
telemetry sender records — so the package also proves the pipeline detects an
arbitrary third-party skill, not only the sender's own self-test.
