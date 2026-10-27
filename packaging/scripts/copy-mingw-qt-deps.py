#!/usr/bin/env python3
"""Copy MinGW/Qt transitive DLLs next to a Windows deploy tree.

windeployqt ships Qt6*.dll + plugins but not ICU/zlib/pcre/freetype/etc.
Also recursively pulls PE imports from /ucrt64/bin.
"""
from __future__ import annotations

import argparse
import re
import shutil
import subprocess
import sys
from pathlib import Path

DLL_RE = re.compile(r"DLL Name:\s+(\S+\.dll)", re.I)
SYS_DLL = {
    x.lower()
    for x in """
    kernel32.dll user32.dll gdi32.dll shell32.dll advapi32.dll ole32.dll
    oleaut32.dll ws2_32.dll winmm.dll imm32.dll version.dll dwmapi.dll
    uxtheme.dll ntdll.dll opengl32.dll d3d11.dll dxgi.dll dwrite.dll
    comdlg32.dll comctl32.dll netapi32.dll userenv.dll crypt32.dll bcrypt.dll
    secur32.dll iphlpapi.dll mpr.dll dnsapi.dll authz.dll shlwapi.dll
    shcore.dll msvcrt.dll ucrtbase.dll nsi.dll win32u.dll rpcrt4.dll
    sechost.dll combase.dll setupapi.dll wtsapi32.dll d3d12.dll dbghelp.dll
    wsock32.dll wininet.dll winhttp.dll normaliz.dll dhcpcsvc.dll
    """.split()
}

PATTERNS = [
    "Qt6Core.dll", "Qt6Gui.dll", "Qt6Widgets.dll", "Qt6Network.dll", "Qt6OpenGL.dll",
    "Qt6Qml.dll", "Qt6QmlModels.dll", "Qt6QmlWorkerScript.dll", "Qt6QmlMeta.dll",
    "Qt6Quick.dll", "Qt6QuickControls2.dll", "Qt6QuickControls2Impl.dll",
    "Qt6QuickLayouts.dll", "Qt6QuickTemplates2.dll",
    "Qt6QuickDialogs2.dll", "Qt6QuickDialogs2QuickImpl.dll", "Qt6QuickDialogs2Utils.dll",
    "Qt6Svg.dll", "Qt6SvgWidgets.dll",
    "libgcc_s_seh-1.dll", "libstdc++-6.dll", "libwinpthread-1.dll",
    "libb2-1.dll", "libbrotlicommon.dll", "libbrotlicommon-1.dll",
    "libbrotlidec.dll", "libbrotlidec-1.dll", "libbz2-1.dll",
    "libdouble-conversion.dll", "libfreetype-6.dll", "libglib-2.0-0.dll",
    "libgraphite2.dll", "libgraphite2-3.dll", "libharfbuzz-0.dll", "libiconv-2.dll",
    "libicudt*.dll", "libicuin*.dll", "libicuuc*.dll", "libicutu*.dll", "libiculx*.dll",
    "icudt*.dll", "icuin*.dll", "icuuc*.dll",
    "libintl-8.dll", "libmd4c.dll", "libpcre2-16-0.dll", "libpcre2-8-0.dll",
    "libpng16-16.dll", "libzstd.dll", "zlib1.dll",
    "libssl-3-x64.dll", "libcrypto-3-x64.dll",
    "libEGL.dll", "libGLESv2.dll", "opengl32sw.dll", "D3Dcompiler_47.dll",
]

REQUIRED = [
    "mclauncher.exe",
    "libgcc_s_seh-1.dll",
    "libstdc++-6.dll",
    "libwinpthread-1.dll",
    "platforms/qwindows.dll",
    "libpcre2-16-0.dll",
    "zlib1.dll",
    "libzstd.dll",
    "libdouble-conversion.dll",
    "libb2-1.dll",
    "libfreetype-6.dll",
    "libharfbuzz-0.dll",
    "libmd4c.dll",
    "libpng16-16.dll",
]


def imports(pe: Path) -> list[str]:
    try:
        out = subprocess.check_output(
            ["objdump", "-p", str(pe)],
            stderr=subprocess.DEVNULL,
            text=True,
            errors="replace",
        )
    except Exception:
        return []
    return [m.group(1) for m in DLL_RE.finditer(out)]


def copy_patterns(ucrt: Path, dest: Path) -> int:
    n = 0
    for pattern in PATTERNS:
        for src in ucrt.glob(pattern):
            if not src.is_file():
                continue
            target = dest / src.name
            shutil.copy2(src, target)
            n += 1
    return n


def copy_recursive(ucrt: Path, dest: Path) -> int:
    index = {p.name.lower(): p for p in ucrt.glob("*.dll")}
    n = 0
    changed, rounds = True, 0
    while changed and rounds < 25:
        changed = False
        rounds += 1
        for pe in list(dest.rglob("*.dll")) + list(dest.rglob("*.exe")):
            for name in imports(pe):
                low = name.lower()
                if low in SYS_DLL:
                    continue
                if (dest / name).exists() or (dest / Path(name).name).exists():
                    continue
                # also accept already-copied case variants
                if any(p.name.lower() == low for p in dest.glob("*.dll")):
                    continue
                src = index.get(low)
                if src is None:
                    continue
                print(f"copy missing dep {src.name}")
                shutil.copy2(src, dest / src.name)
                n += 1
                changed = True
    return n


def needs_icu(dest: Path) -> bool:
    core = dest / "Qt6Core.dll"
    if not core.exists():
        return False
    return any("icu" in n.lower() for n in imports(core))


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("dest", type=Path, help="deploy directory")
    ap.add_argument("--ucrt", type=Path, default=Path("/ucrt64/bin"))
    args = ap.parse_args()
    dest: Path = args.dest
    ucrt: Path = args.ucrt
    if not dest.is_dir():
        print(f"ERROR: dest not a directory: {dest}", file=sys.stderr)
        return 1
    if not ucrt.is_dir():
        print(f"ERROR: ucrt bin not found: {ucrt}", file=sys.stderr)
        return 1

    print(f"Copying MinGW/Qt dependency DLLs from {ucrt}")
    n = copy_patterns(ucrt, dest)
    print(f"pattern copies: {n}")
    n2 = copy_recursive(ucrt, dest)
    print(f"recursive copies: {n2}")

    qtconf = dest / "qt.conf"
    if not qtconf.exists():
        qtconf.write_text(
            "[Paths]\nPrefix=.\nPlugins=.\nQmlImports=qml\nLibraries=.\n",
            encoding="utf-8",
        )

    miss = 0
    for rel in REQUIRED:
        p = dest / rel
        if not p.is_file():
            print(f"ERROR: missing required file: {p}", file=sys.stderr)
            miss = 1

    if needs_icu(dest):
        families = {
            "icuuc": list(dest.glob("libicuuc*.dll")) + list(dest.glob("icuuc*.dll")),
            "icuin": list(dest.glob("libicuin*.dll")) + list(dest.glob("icuin*.dll")),
            "icudt": list(dest.glob("libicudt*.dll")) + list(dest.glob("icudt*.dll")),
        }
        for fam, hits in families.items():
            if not hits:
                print(f"ERROR: Qt6Core needs ICU but no {fam} DLL in {dest}", file=sys.stderr)
                miss = 1
        if miss:
            print("---- ucrt ICU candidates ----", file=sys.stderr)
            for p in sorted(ucrt.glob("*icu*")):
                print(f"  {p}", file=sys.stderr)
    else:
        print("Qt6Core does not import ICU; skipping ICU assert")

    if miss:
        print(f"---- {dest} dlls ----", file=sys.stderr)
        for p in sorted(dest.glob("*.dll")):
            print(f"  {p.name}", file=sys.stderr)
        return 1

    print("MinGW/Qt dependency copy OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
