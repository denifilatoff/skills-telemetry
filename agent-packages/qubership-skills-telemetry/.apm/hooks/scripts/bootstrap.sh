#!/bin/sh
# Locate-or-download the skills-telemetry binary into a per-machine cache,
# then exec it forwarding all args and stdin. POSIX sh; only built-in tools.
#
# Two callers share this script:
#   - the hook, as `bootstrap.sh ingest --agent=...` — must never fail the
#     agent turn, so bootstrap's own errors exit 0;
#   - the provisioning one-liner, as `bootstrap.sh provision` — interactive, so
#     bootstrap's own errors must surface with a non-zero exit.
set -eu

# BINARY_VERSION is the release tag; BASE_URL/<tag>/<asset> is the GitHub
# Releases download layout, so the URL below resolves to a real asset.
BINARY_VERSION="v0.5.2"
BASE_URL="https://github.com/denifilatoff/skills-telemetry/releases/download"

CMD="${1:-}"

# die surfaces a bootstrap failure. The hook path (ingest) swallows it so the
# agent turn is never broken; every other command exits non-zero.
die() {
  echo "skills-telemetry: $1" >&2
  [ "$CMD" = "ingest" ] && exit 0
  exit 1
}

# Per-OS cache base (mirrors Go os.UserCacheDir).
case "$(uname -s)" in
  Darwin) CACHE_BASE="$HOME/Library/Caches" ;;
  *)      CACHE_BASE="${XDG_CACHE_HOME:-$HOME/.cache}" ;;
esac

case "$(uname -s)" in
  Darwin) OS="darwin" ;;
  Linux)  OS="linux" ;;
  *) die "unsupported OS $(uname -s)" ;;
esac
case "$(uname -m)" in
  arm64|aarch64) ARCH="arm64" ;;
  x86_64|amd64)  ARCH="amd64" ;;
  *) die "unsupported arch $(uname -m)" ;;
esac

DIR="$CACHE_BASE/qubership-skills-telemetry/bin/$BINARY_VERSION"
BIN="$DIR/skills-telemetry-$OS-$ARCH"

if [ ! -x "$BIN" ]; then
  mkdir -p "$DIR"
  TMP="$BIN.tmp.$$"
  if ! curl -fsSL "$BASE_URL/$BINARY_VERSION/skills-telemetry-$OS-$ARCH" -o "$TMP"; then
    rm -f "$TMP"
    die "download failed"
  fi
  chmod +x "$TMP"
  mv "$TMP" "$BIN"
fi

exec "$BIN" "$@"
