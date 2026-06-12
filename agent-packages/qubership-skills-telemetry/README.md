# qubership-skills-telemetry

This package provides target-specific hook implementations for observing
visible skill invocation breadcrumbs.

Current status:

- Codex: implemented fully
- Cursor: placeholder
- Claude: placeholder

The Codex `Stop` hook calls `bootstrap.sh` (macOS/Linux) or `bootstrap.ps1`
(Windows). The bootstrap fetches the pinned `skills-telemetry` Go binary into
a per-machine cache on first run, then invokes `ingest` with `--agent=codex`
and the configured `--endpoint`. The `ingest` command reads the hook payload
from stdin, normalizes the event, and writes it to a machine-global spool.
The same ingest run opportunistically flushes buffered events to the collector over OTLP/HTTP (no daemon).

Codex is implemented. Claude, Cursor, and OpenCode adapters are follow-up work.

## Release

1. Bump `VERSION` in `sender/Makefile` (default `0.1.0`).
2. From the `sender/` directory, run `make checksums`. This cross-compiles five
   binaries into `sender/dist/` and writes `sender/dist/SHA256SUMS`.
3. Upload every file in `sender/dist/` to the artifact store (GitHub Releases,
   S3, or equivalent).
4. In `bootstrap.sh` and `bootstrap.ps1`, set:
   - `BINARY_VERSION` to the version you bumped in step 1.
   - `BASE_URL` to the base URL of the artifact store where you uploaded the binaries.
   - The per-platform SHA-256 values to the corresponding lines from `SHA256SUMS`.

Checksum verification in the bootstrap scripts is added once the first real
release artifacts are in place.
