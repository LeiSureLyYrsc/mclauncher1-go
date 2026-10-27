#!/usr/bin/env bash
# Build a pure-Go SFX stub and append a zip of DIR as trailing payload.
#
# Usage: make-sfx.sh <deploy-dir> <output-path> [inner-binary-name]
# Example: make-sfx.sh dist/mclauncher-windows-amd64 dist/mclauncher-windows-amd64.exe mclauncher.exe
set -eu

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SFX_DIR="$ROOT/packaging/sfx"

DEPLOY_DIR="${1:?deploy dir}"
OUT="${2:?output path}"
INNER="${3:-}"

if [ ! -d "$DEPLOY_DIR" ]; then
  echo "ERROR: deploy dir not found: $DEPLOY_DIR" >&2
  exit 1
fi

# Detect inner binary if not given.
if [ -z "$INNER" ]; then
  if [ -f "$DEPLOY_DIR/mclauncher.exe" ]; then
    INNER=mclauncher.exe
  elif [ -f "$DEPLOY_DIR/mclauncher" ]; then
    INNER=mclauncher
  else
    echo "ERROR: neither mclauncher.exe nor mclauncher found in $DEPLOY_DIR" >&2
    exit 1
  fi
fi
if [ ! -f "$DEPLOY_DIR/$INNER" ]; then
  echo "ERROR: missing $DEPLOY_DIR/$INNER" >&2
  exit 1
fi

TMP="$(mktemp -d)"
cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT

STUB="$TMP/stub"
PAYLOAD="$TMP/payload.zip"

echo "Building SFX stub (CGO_ENABLED=0)..."
(
  cd "$SFX_DIR"
  # On Windows, `go build -o stub` may still emit stub.exe — accept either.
  CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -o "$STUB" .
)
if [ ! -f "$STUB" ] && [ -f "${STUB}.exe" ]; then
  STUB="${STUB}.exe"
fi
if [ ! -f "$STUB" ]; then
  echo "ERROR: SFX stub not produced at $STUB" >&2
  exit 1
fi

echo "Zipping deploy dir (contents at zip root)..."
# Zip contents of deploy dir so extract yields mclauncher(.exe) at cache root.
(
  cd "$DEPLOY_DIR"
  # Prefer zip; fall back to Go archive if needed.
  if command -v zip >/dev/null 2>&1; then
    zip -q -r "$PAYLOAD" .
  else
    go run "$ROOT/packaging/scripts/zipdir.go" "$PAYLOAD" .
  fi
)

PAYLOAD_SIZE=$(wc -c < "$PAYLOAD" | tr -d ' ')
echo "Payload size: $PAYLOAD_SIZE bytes"

mkdir -p "$(dirname "$OUT")"
cat "$STUB" "$PAYLOAD" > "$OUT"

# Append uint64 LE size + magic "MCLSFX01"
python3 - "$OUT" "$PAYLOAD_SIZE" <<'PY'
import struct, sys
out, size = sys.argv[1], int(sys.argv[2])
with open(out, "ab") as f:
    f.write(struct.pack("<Q", size))
    f.write(b"MCLSFX01")
PY

chmod +x "$OUT" 2>/dev/null || true
echo "Wrote SFX: $OUT ($(wc -c < "$OUT" | tr -d ' ') bytes)"
