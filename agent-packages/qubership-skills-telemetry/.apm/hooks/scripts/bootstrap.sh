#!/bin/sh
# Locate-or-download the skills-telemetry binary into a per-machine cache,
# then exec it forwarding all args and stdin. POSIX sh; only built-in tools.
set -eu

# BINARY_VERSION is the release tag; BASE_URL/<tag>/<asset> is the GitHub
# Releases download layout, so the URL below resolves to a real asset.
BINARY_VERSION="v0.1.1"
BASE_URL="https://github.com/denifilatoff/skills-telemetry/releases/download"

# Per-OS cache base (mirrors Go os.UserCacheDir).
case "$(uname -s)" in
  Darwin) CACHE_BASE="$HOME/Library/Caches" ;;
  *)      CACHE_BASE="${XDG_CACHE_HOME:-$HOME/.cache}" ;;
esac

case "$(uname -s)" in
  Darwin) OS="darwin" ;;
  Linux)  OS="linux" ;;
  *) echo "skills-telemetry: unsupported OS $(uname -s)" >&2; exit 0 ;;
esac
case "$(uname -m)" in
  arm64|aarch64) ARCH="arm64" ;;
  x86_64|amd64)  ARCH="amd64" ;;
  *) echo "skills-telemetry: unsupported arch $(uname -m)" >&2; exit 0 ;;
esac

DIR="$CACHE_BASE/qubership-skills-telemetry/bin/$BINARY_VERSION"
BIN="$DIR/skills-telemetry-$OS-$ARCH"

if [ ! -x "$BIN" ]; then
  mkdir -p "$DIR"
  TMP="$BIN.tmp.$$"
  if ! curl -fsSL "$BASE_URL/$BINARY_VERSION/skills-telemetry-$OS-$ARCH" -o "$TMP"; then
    echo "skills-telemetry: download failed" >&2; exit 0
  fi
  chmod +x "$TMP"
  mv "$TMP" "$BIN"
fi

exec "$BIN" "$@"
