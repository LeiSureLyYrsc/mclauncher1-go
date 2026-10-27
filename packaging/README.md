# Single-file packaging (Qt6 + QML)

## Why not Enigma / BoxedApp / UPX virtual FS?

Qt Quick loads **QML modules and plugins from a real directory tree**
(`platforms/qwindows.dll`, `qml/QtQuick/...`). Virtual-filesystem packers
(Enigma Virtual Box, BoxedApp Packer, etc.) frequently break QML import
resolution or plugin discovery. UPX alone only compresses the PE and does
**not** embed sibling DLLs.

## Chosen approach: deploy dir + Go SFX stub

1. **Build** the miqt/Qt6 binary normally (`CGO_ENABLED=1`).
2. **Deploy runtime next to it**
   - Windows: `windeployqt --qmldir qml`, then **always** copy MinGW runtime
     (`libgcc` / `libstdc++` / `libwinpthread`) **and** Qt transitive deps
     (ICU / zlib / pcre2 / freetype / harfbuzz / …). `windeployqt` alone does
     not ship those on MSYS2; missing them makes clean PCs fail to start.
     Also write `qt.conf` (`Prefix=.`, `Plugins=.`, `QmlImports=qml`).
   - Linux: `packaging/scripts/bundle-linux-qt.sh` copies the `.so` closure,
     `plugins/`, and `qml/` modules, writes `qt.conf`.
3. **Wrap with a pure-Go self-extractor** (`packaging/sfx`):
   - Zip the deploy directory.
   - Append zip + `uint64` size + magic `MCLSFX01` to the stub.
   - On run: extract once to `%TEMP%/mclauncher-sfx-<hash>/` (or `/tmp/...`),
     set `QT_PLUGIN_PATH` / `QT_QPA_PLATFORM_PLUGIN_PATH` / `QML_IMPORT_PATH` /
     `PATH` or `LD_LIBRARY_PATH`, exec the real `mclauncher(.exe)` with CWD =
     extract dir (launch dir passed as `MCLAUNCHER_CWD`). Failures write a
     log and show a MessageBox on Windows.

User downloads **one file**. First launch unpacks Qt (~tens of MB); later
launches reuse the cache.

Optional zip of the deploy directory is still produced for debugging.

## Local usage

```bash
# After you have dist/mclauncher-windows-amd64/ (or linux) populated:
./packaging/scripts/make-sfx.sh \
  dist/mclauncher-windows-amd64 \
  dist/mclauncher-windows-amd64.exe \
  mclauncher.exe

./packaging/scripts/bundle-linux-qt.sh ./mclauncher dist/mclauncher-linux-amd64
./packaging/scripts/make-sfx.sh \
  dist/mclauncher-linux-amd64 \
  dist/mclauncher-linux-amd64 \
  mclauncher
```

## CI artifacts

| Platform | Primary (single file)              | Optional          |
|----------|------------------------------------|-------------------|
| Windows  | `mclauncher-windows-amd64.exe`     | `.zip` of deploy  |
| Linux    | `mclauncher-linux-amd64` (SFX)     | `.tar.gz` deploy  |
