# Install

You install the package with APM and let its setup skill do the rest. The skill
asks what it needs and configures the machine for you — you don't run the binary
by hand.

## 1. Install the package

Add the package to your `apm.yml` dependencies:

```yaml
dependencies:
  apm:
    - denifilatoff/skills-telemetry/agent-packages/qubership-skills-telemetry#v0.5.0
```

Then install and compile for your agent (swap `codex` for your target):

```sh
apm install --target codex
apm compile --target codex
```

This deploys the setup skill and the telemetry hook, and merges the skill trigger
into `AGENTS.md`. Installing is the consent boundary — nothing is sent until the
skill provisions a collector endpoint.

## 2. Run the setup skill

In your agent, ask it to set up telemetry — for example "set up skills telemetry"
or "provision telemetry". This runs the `provision-skills-telemetry` skill, which:

- asks for the collector endpoint, and the CA certificate where the collector uses
  a private one;
- prompts for a token only if the collector needs one, read without echo;
- writes the config, then verifies the pipeline with a live probe.

The skill reports success only after a probe reaches the collector. There is no
manual step beyond answering its questions.

## Headless / CI

Without an agent, run the same setup from the shell:

```sh
curl -fsSL https://github.com/denifilatoff/skills-telemetry/releases/latest/download/bootstrap.sh | sh -s -- provision
```

## Reference: binary commands

The setup skill calls these for you; you rarely run them by hand. `skills-telemetry
<command>`:

| Command | Purpose |
|---|---|
| `provision` | Write the per-machine config: collector endpoint, optional CA certificate (`--ca=<path>`), and an optional token read without echo. Idempotent. |
| `status` | Read-only check: build version, config directory, endpoint, whether a CA is present, spool backlog, last flush attempt, and a provisioned verdict. Sends nothing. |
| `selftest` | Send one marked probe event and report whether the collector accepted it and it left the spool. |
| `ingest` | The hook path: read an agent hook payload on stdin, detect skill use (on Codex the `[skill-called]` marker plus the `SKILL.md` reads in the session rollout; on Claude Code the `Skill` tool name in the `PreToolUse` payload; on Cursor the marker plus the `SKILL.md` reads in the `afterAgentResponse` transcript), queue the events, and flush opportunistically. Always exits 0 so it never fails an agent turn. |
| `flush` | Send queued events to the collector and delete each on success. |
| `version` | Print the build version. |
