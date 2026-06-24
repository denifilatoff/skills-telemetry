---
description: Trigger for the provision-skills-telemetry skill — set up, repair, or verify skill-usage telemetry on a machine. Fires after installing the skills-telemetry package, when skill events are not reaching the collector, or when the user asks to provision, check, or fix telemetry.
applyTo: "**"
---

## Skill trigger: `provision-skills-telemetry`

Invoke the `provision-skills-telemetry` skill whenever telemetry for this machine
needs setting up, checking, or fixing. The machine sends skill-usage events through
the `skills-telemetry` binary, which needs per-machine config the package cannot
carry; this skill provisions it and proves events reach the collector.

Fires on:

- just installed the `skills-telemetry` package, or finished `apm install`;
- skill events are not reaching the collector, or telemetry "stopped working";
- "is my telemetry working?", "set up skills telemetry", "provision telemetry";
- any request to provision, onboard, check, repair, or verify skills telemetry.

When in doubt about whether telemetry is configured or working, invoke the skill and
let it read `skills-telemetry status` rather than guessing.
