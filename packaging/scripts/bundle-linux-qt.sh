#!/usr/bin/env bash
# Bundle a dynamically linked Qt6/miqt Linux binary with its .so deps, plugins, and QML modules.
#
# Usage: bundle-linux-qt.sh <mclauncher-binary> <dest-dir>
# Produces dest-dir/mclauncher plus libs/plugins/qml suitable for the SFX stub.
set -eu

BIN="${1:?binary}"
DEST="${2:?dest dir}"

if [ ! -f "$BIN" ]; then
  echo "ERROR: binary not found: $BIN" >&2
  exit 1
fi

rm -rf "$DEST"
mkdir -p "$DEST"
cp -f "$BIN" "$DEST/mclauncher"
chmod +x "$DEST/mclauncher"

# --- collect shared library closure (skip glibc / loader / mesa GL) ---
is_system_lib() {
  case "$1" in
    */ld-linux*.so*|*/libc.so*|*/libm.so*|*/libdl.so*|*/librt.so*|*/libpthread.so*|*/libresolv.so*|*/libutil.so*)
      return 0 ;;
    */libstdc++.so*|*/libgcc_s.so*)
      # Bundle libstdc++/libgcc for older target hosts — comment to rely on system.
      return 1 ;;
    */libGL.so*|*/libOpenGL.so*|*/libGLdispatch.so*|*/libGLX.so*|*/libEGL.so*)
      return 0 ;;
    */libglib-2.0.so*|*/libgobject-2.0.so*|*/libgio-2.0.so*|*/libgmodule-2.0.so*|*/libgthread-2.0.so*)
      return 0 ;;
    */libX11.so*|*/libxcb*.so*|*/libXext.so*|*/libXau.so*|*/libXdmcp.so*|*/libXrender.so*|*/libXi.so*|*/libXcursor.so*|*/libXfixes.so*|*/libXrandr.so*|*/libXinerama.so*)
      return 0 ;;
    */libfontconfig.so*|*/libfreetype.so*|*/libexpat.so*|*/libz.so*|*/libbz2.so*|*/libpng*.so*|*/libjpeg*.so*)
      return 0 ;;
    */libdbus-1.so*|*/libsystemd.so*|*/libudev.so*|*/libselinux.so*|*/libmount.so*|*/libblkid.so*|*/libpcre*.so.3*)
      return 0 ;;
    */libdrm.so*|*/libwayland*.so*|*/libffi.so*)
      return 0 ;;
    *)
      return 1 ;;
  esac
}

echo "Collecting shared library closure..."
declare -A SEEN=()
QUEUE=("$DEST/mclauncher")
while [ ${#QUEUE[@]} -gt 0 ]; do
  CUR="${QUEUE[0]}"
  QUEUE=("${QUEUE[@]:1}")
  [ -f "$CUR" ] || continue
  while read -r lib; do
    [ -n "$lib" ] || continue
    [ -e "$lib" ] || continue
    real="$(readlink -f "$lib" || echo "$lib")"
    base="$(basename "$real")"
    if [ -n "${SEEN[$base]:-}" ]; then
      continue
    fi
    if is_system_lib "$real"; then
      continue
    fi
    SEEN[$base]=1
    cp -L "$real" "$DEST/$base"
    QUEUE+=("$DEST/$base")
  done < <(ldd "$CUR" 2>/dev/null | awk '/=>/ {print $3} /^[^[:space:]]/ { if ($0 ~ /ld-linux/) next }')
done

# --- Qt plugins ---
QT_PLUGINS=""
for cand in \
  /usr/lib/x86_64-linux-gnu/qt6/plugins \
  /usr/lib/qt6/plugins \
  /usr/lib64/qt6/plugins; do
  if [ -d "$cand" ]; then
    QT_PLUGINS="$cand"
    break
  fi
done

if [ -n "$QT_PLUGINS" ]; then
  echo "Copying Qt plugins from $QT_PLUGINS"
  mkdir -p "$DEST/plugins"
  for plugindir in platforms xcbglintegrations styles imageformats tls networkinformation iconengines \
                   qmltooling generic platforminputcontexts; do
    if [ -d "$QT_PLUGINS/$plugindir" ]; then
      mkdir -p "$DEST/plugins/$plugindir"
      cp -a "$QT_PLUGINS/$plugindir"/. "$DEST/plugins/$plugindir/" 2>/dev/null || true
    fi
  done
else
  echo "WARNING: Qt6 plugins directory not found" >&2
fi

# --- QML modules ---
QT_QML=""
for cand in \
  /usr/lib/x86_64-linux-gnu/qt6/qml \
  /usr/lib/qt6/qml \
  /usr/lib64/qt6/qml; do
  if [ -d "$cand" ]; then
    QT_QML="$cand"
    break
  fi
done

if [ -n "$QT_QML" ]; then
  echo "Copying QML modules from $QT_QML"
  mkdir -p "$DEST/qml"
  for mod in QtQml QtQuick QML Qt; do
    if [ -d "$QT_QML/$mod" ]; then
      cp -a "$QT_QML/$mod" "$DEST/qml/"
    fi
  done
else
  echo "WARNING: Qt6 QML directory not found" >&2
fi

# Also harvest .so files referenced by plugins/QML plugins into DEST.
echo "Collecting plugin/QML .so dependencies..."
while read -r so; do
  [ -f "$so" ] || continue
  while read -r lib; do
    [ -n "$lib" ] || continue
    [ -e "$lib" ] || continue
    real="$(readlink -f "$lib" || echo "$lib")"
    base="$(basename "$real")"
    if [ -n "${SEEN[$base]:-}" ]; then
      continue
    fi
    if is_system_lib "$real"; then
      continue
    fi
    SEEN[$base]=1
    cp -L "$real" "$DEST/$base"
  done < <(ldd "$so" 2>/dev/null | awk '/=>/ {print $3}')
done < <(find "$DEST/plugins" "$DEST/qml" -type f -name '*.so*' 2>/dev/null || true)

# qt.conf so Qt finds plugins relative to the binary even without env vars.
cat > "$DEST/qt.conf" <<'QTC'
[Paths]
Prefix=.
Plugins=plugins
QmlImports=qml
Qml2Imports=qml
QTC

echo "Bundled files: $(find "$DEST" -type f | wc -l)"
ls -la "$DEST/mclauncher"
ls -la "$DEST/plugins/platforms" 2>/dev/null || echo "(no platforms plugins)"
ls -la "$DEST/qml" 2>/dev/null || echo "(no qml)"
# Spot-check a Qt lib was copied
ls "$DEST"/libQt6Core.so* 2>/dev/null || ls "$DEST"/libQt6Qml.so* 2>/dev/null || \
  echo "WARNING: no libQt6Core/Qml in bundle — binary may already be mostly static or ldd missed Qt" >&2
