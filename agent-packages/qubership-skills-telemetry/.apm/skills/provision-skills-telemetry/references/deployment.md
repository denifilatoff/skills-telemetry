# Deployment specifics

These values depend on where the collector runs, so the skill stays generic and asks for them
rather than hardcoding any one deployment.

## Fetch the binary

Run once per machine to place the binary in the per-machine cache and start provisioning:

```sh
# macOS / Linux
curl -fsSL https://github.com/denifilatoff/skills-telemetry/releases/latest/download/bootstrap.sh | sh -s -- provision
# Windows (PowerShell)
iex "& { $(irm https://github.com/denifilatoff/skills-telemetry/releases/latest/download/bootstrap.ps1) } provision"
```

The bootstrap scripts are published as release assets, so `releases/latest/download` always
resolves to the current launcher.

## Endpoint

The OTLP/HTTP logs endpoint, of the form `https://<collector-host>/v1/logs`. Get it from the
onboarding portal or an admin. Always `https://` — the sender never sends over plaintext.

## CA certificate

Needed only when the collector's certificate does not chain to a root the machine already
trusts.

- **Public certificate or MDM-distributed corporate CA** — already in the system trust store;
  nothing to do.
- **Corporate CA, not in the trust store** — download the root CA from the internal
  distribution point, then pass its path to `provision --ca`.
- **Local self-signed cluster (cert-manager)** — extract the CA from the issuing secret, then
  pass the file to `provision --ca`:

  ```sh
  kubectl -n <namespace> get secret <ca-secret> \
    -o jsonpath='{.data.tls\.crt}' | base64 -d > ca.crt
  ```

## Confirm delivery in the store (optional)

Only if the user has read access to the store. After `selftest`, confirm the probe landed by
querying for its probe name, e.g. against VictoriaLogs:

```sh
curl -s '<query-url>/select/logsql/query' --data-urlencode 'query=skill.name:="__selftest__"'
```

The probe carries `skill.name="__selftest__"`, so it is easy to find and easy to filter out of
real skill-usage metrics on the collector.

Most participants won't have read access — a passing `selftest` (accepted and dequeued) is the
guarantee to rely on.
