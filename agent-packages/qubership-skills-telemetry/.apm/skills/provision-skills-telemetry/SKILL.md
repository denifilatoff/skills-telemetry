---
name: provision-skills-telemetry
description: Set up, repair, and verify Qubership skill-usage telemetry on a machine. Use right after installing the qubership-skills-telemetry package, when skill events are not reaching the collector, when telemetry "stopped working", or whenever the user asks to provision, onboard, check, or fix skills telemetry — even phrased loosely as "is my telemetry working?" or "set up skills telemetry".
---

# Provision skills telemetry

This machine reports skill-usage telemetry through a small binary, `skills-telemetry`.
The binary needs per-machine config the public package cannot carry: a collector endpoint,
sometimes a CA certificate, sometimes a token. Your job is to get that config in place and
prove events actually reach the collector — then stop.

You orchestrate; the binary does the sensitive work. It owns the config files (atomic writes,
permissions, idempotency) and reads the token without echo. Discover and ask; let the binary
write. Never put the token in your own output.

## What "working" means

- `skills-telemetry status` — read-only state: binary version, config dir, endpoint, whether a
  CA file is present, spool backlog, last flush attempt, and a provisioned/not verdict.
- `skills-telemetry selftest` — sends one real, marked probe event and reports whether the
  collector accepted it and the event left the spool.
- Config lives under the config dir that `status` prints: `env` (endpoint, token) and an
  optional `ca.crt`. These are the binary's to write — don't hand-edit them.

## Workflow

Read state first, close only the gaps it shows, then prove delivery.

1. Run `status`. If the binary is missing, fetch it with the bootstrap one-liner
   (`references/deployment.md`), then retry.
2. Fix each gap `status` reports (next section).
3. Run `selftest`. Re-run `status` / `selftest` after each fix until it passes.
4. Report the outcome (see "Verify delivery").

## If you were handed a config file

When the user points you to a ready `env` file (it carries `SKILLS_TELEMETRY_ENDPOINT` and
`SKILLS_TELEMETRY_TOKEN`), copy it into place instead of provisioning field by field:

1. Read the config dir from `status` (the `config_dir:` line).
2. Copy the file there as `env`, verbatim:
   `mkdir -p <config_dir> && cp <given-file> <config_dir>/env`.
3. Run `selftest` to confirm delivery.

Do not open, read, print, or echo the file — it holds the token, and anything in this
conversation enters the model's context. A copy moves the bytes without reading them. The
CLI mints the anonymous `machine-id` itself on first send, so the two properties are
enough.

## Closing gaps

- **Endpoint missing** — ask the user for the collector URL; their onboarding portal or admin
  has it. Run `skills-telemetry provision --endpoint=<url>`.
- **CA needed** (`selftest` fails with a certificate / TLS error) — only self-signed or
  non-trusted-CA deployments need this; a publicly trusted or MDM-distributed CA needs nothing.
  Obtain the `.crt` (`references/deployment.md` covers a local cluster and a corporate PKI) and
  run `skills-telemetry provision --ca=<path>`; the binary copies it to `ca.crt`.
- **Token required** (collector returns 401 / 403) — have the user type it into the binary's
  no-echo prompt: run `skills-telemetry provision` and let them enter the token when asked.
  Don't ask the user to paste the token to you, and don't type it yourself — anything in this
  conversation becomes part of the model's context and would leak the secret.

## Failure → fix

| `status` / `selftest` shows | Cause | Fix |
|---|---|---|
| binary not found | not fetched yet | run the bootstrap one-liner |
| endpoint empty | not provisioned | `provision --endpoint` |
| TLS verification failed | CA missing or wrong | `provision --ca` |
| connection refused / timeout | network or VPN | confirm the user can reach the collector host |
| 401 / 403 | token missing or rejected | `provision`, enter the token at the no-echo prompt |
| spool growing, flush failing | one of the above | fix the reported cause, then `selftest` |
| Cursor: `selftest` passes but real skill runs send nothing | `.cursor/hooks.json` lost its top-level `version` on a fresh `apm install` | add `"version": 1` (see "Cursor: confirm the hook fires") |

`selftest` prints the raw send error (for example an `x509` / `tls` message or an HTTP status);
map it to a cause above. `status` shows the spool backlog and the provisioned/not verdict but
does not itself test the network.

## Cursor: confirm the hook fires

On Cursor only, check that `.cursor/hooks.json` has a numeric top-level `version`. A fresh
`apm install` drops it, and without it Cursor silently loads no hooks — so skills run but
nothing ever reaches `ingest`. `selftest` still passes, because it calls the binary directly,
so the broken hook hides behind a green check: the sender works, yet real skill events are
never captured.

If the field is missing, add it:

```json
{
  "version": 1,
  "hooks": { ... }
}
```

A reinstall over a file that already has `version` keeps it; only a fresh install drops it.

This is a workaround for an APM bug, tracked at https://github.com/microsoft/apm/issues/1823.
Once that issue ships a fix, this step and the matching row in "Failure → fix" are obsolete —
delete them.

## Verify delivery

`selftest` sends a real event as a test. Two outcomes count as success:

- The collector accepted it (HTTP 200) and it left the spool — the pipeline works end to end up
  to ingest. This is the guarantee you can always make.
- If the user has read access to the store (VictoriaLogs or similar), offer the query that
  confirms the probe landed (`references/deployment.md`). Most users won't have it — don't block
  on it.

If the probe stays in the spool, delivery failed: treat it as a gap and diagnose from `status`.

Don't report success without a passing `selftest`. A written config that can't reach the
collector looks done but sends nothing.
