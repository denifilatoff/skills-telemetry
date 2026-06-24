# Codex sandbox: let the hook out (execpolicy rule)

Reference for the "Codex sandbox rule" check in `SKILL.md`. Read this when Codex is one of the
targets and the rule is missing or inert, and you need to write or repair it.

## Why this is needed

Codex runs hook commands and tool calls in a **sandbox** that, by default, denies the binary the
two things telemetry needs: read access to the machine-level config outside the project
(`~/.config/skills-telemetry/`) and network egress to the collector. The tell is a
**Codex-only** failure: inside Codex, `status` reports `endpoint: (unset)` / `not provisioned` and
`selftest` fails with `no endpoint`, while the *same* binary run from Claude Code or a plain shell
reports `provisioned` and delivers; `update-check` reporting `latest: unknown` is the same sandbox
blocking egress to GitHub. This is **not** a missing provision — the config is on disk; the sandbox
just hides it. (It is also not the MSIX `%AppData%` issue, which the `~/.config` move already
fixed.)

The fix is a Codex **execution-policy rule** that lets those telemetry commands run outside the
sandbox. Keep it machine-level, like the rest of telemetry (`~/.local/bin`, `~/.config`): write it
to the **user layer** `~/.codex/rules/`, which Codex loads in every repo with no per-project trust
step.

## Write the rule

Write `~/.codex/rules/skills-telemetry.rules` (on Windows
`%USERPROFILE%\.codex\rules\skills-telemetry.rules`) with exactly the content below. Unlike `env`
this is a static policy, not a secret, so you write it directly. It allows only the three commands
that must leave the sandbox — `ingest --agent=codex`, `status`, `selftest` — and deliberately
leaves `provision` sandboxed (it writes config and reads a token at a no-echo prompt; it must not
run unattended outside the sandbox).

```python
# skills-telemetry — allow the telemetry hook + diagnostics out of the Codex sandbox.
# Scoped to these subcommands only; provision stays sandboxed.
prefix_rule(
    pattern = [["skills-telemetry", "skills-telemetry.exe"], "ingest", "--agent=codex"],
    decision = "allow",
    justification = "Allow the trusted telemetry hook to read its machine config and send Codex skill usage events.",
    match = ["skills-telemetry ingest --agent=codex", "skills-telemetry.exe ingest --agent=codex"],
    not_match = ["skills-telemetry status", "skills-telemetry selftest", "skills-telemetry provision",
                 "skills-telemetry update-check", "skills-telemetry ingest --agent=claude",
                 "skills-telemetry ingest --agent=cursor"],
)
prefix_rule(
    pattern = [["skills-telemetry", "skills-telemetry.exe"], "status"],
    decision = "allow",
    justification = "Allow telemetry diagnostics to read provisioned state outside the sandbox.",
    match = ["skills-telemetry status", "skills-telemetry.exe status"],
    not_match = ["skills-telemetry provision", "skills-telemetry selftest",
                 "skills-telemetry ingest --agent=codex", "skills-telemetry update-check"],
)
prefix_rule(
    pattern = [["skills-telemetry", "skills-telemetry.exe"], "selftest"],
    decision = "allow",
    justification = "Allow telemetry diagnostics to send a marked probe event outside the sandbox.",
    match = ["skills-telemetry selftest", "skills-telemetry.exe selftest"],
    not_match = ["skills-telemetry provision", "skills-telemetry status",
                 "skills-telemetry ingest --agent=codex", "skills-telemetry update-check"],
)
```

Codex scans `rules/` at **startup**, so after writing the file Codex must be restarted (a *full*
restart — see "Calling the binary" in `SKILL.md`) before the rule takes effect.

## Check the rule is there and loads

Three checks, cheapest first — run them whenever Codex is a target, not only when something looks
broken, so a missing or inert rule is caught before you report success:

1. **Present** — `~/.codex/rules/skills-telemetry.rules` exists and carries the three
   `prefix_rule`s above.
2. **Valid + allows** — inside Codex (where `codex` is on `PATH`):
   `codex execpolicy check --rules ~/.codex/rules/skills-telemetry.rules "skills-telemetry ingest --agent=codex" --pretty`
   reports `decision: allow` and the matching rule. The `match` / `not_match` lines also self-test
   at load, so a mis-scoped pattern surfaces on Codex startup.
3. **Effective** — from inside Codex, after a restart, call the binary **by its bare name**:
   `skills-telemetry status` shows the real `~/.config` path and `state: provisioned`, and
   `skills-telemetry selftest` delivers.

## Test only with the exact allowlisted forms (the false-negative trap)

execpolicy matches the literal `argv` tokens (prefix only — no substring or regex), and the rule
is keyed to the bare program name: `argv[0]` must be exactly `skills-telemetry` or
`skills-telemetry.exe`, followed by `status`, `selftest`, or `ingest --agent=codex`. Only that
exact shape escapes the sandbox. Call the binary any other way and it runs **inside** the sandbox,
where it cannot read `~/.config` — so it reports a **false** `endpoint: (unset)` / `not provisioned`
/ send failure that says nothing about telemetry, only that you called it the unmatched way. This
is the trap that makes a working install look broken. The misleading forms:

- **Full path or a `&` wrapper** — `& "…\.local\bin\skills-telemetry.exe" status`. `argv[0]` is the
  path, not the bare name, so no rule matches. Use `skills-telemetry status`.
- **A non-allowlisted subcommand** — `version`, `update-check`. The rule deliberately leaves these
  sandboxed, so in Codex `update-check` **always** reports `latest: unknown` — expected, not a
  network fault.
- **Extra or reordered arguments** beyond the exact allowlisted tokens.

So inside Codex, verify only with the bare-name `status` / `selftest`, and never conclude telemetry
is broken from a full-path or non-allowlisted call. Only if the bare-name forms still report
`not provisioned` after a restart is the rule genuinely not taking effect — diagnose that below.

If `execpolicy check` says `allow` but the bare-name `status` still reports `not provisioned`, the
rule is not being loaded: confirm the file sits in the **user** layer `~/.codex/rules/` that Codex
scans. A per-repo `<repo>/.codex/rules/` copy loads only when the project `.codex/` layer is
trusted in Codex — which is exactly why this skill uses the user layer instead. Don't report Codex
telemetry working until check 3 holds.
