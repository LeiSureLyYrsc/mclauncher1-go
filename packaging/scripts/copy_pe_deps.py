#!/usr/bin/env python3
"""Copy missing PE import DLLs from a search dir into dest (recursive)."""
from __future__ import annotations

import argparse
import re
import subprocess
import sys
from pathlib import Path

DLL_RE = re.compile(r"DLL Name:\s+(\S+\.dll)", re.I)
SYS_DLL = {
    x.lower()
    for x in [
        "kernel32.dll",
        "user32.dll",
        "gdi32.dll",
        "shell32.dll",
        "advapi32.dll",
        "ole32.dll",
        "oleaut32.dll",
        "ws2_32.dll",
        "winmm.dll",
        "imm32.dll",
        "version.dll",
        "dwmapi.dll",
        "uxtheme.dll",
        "ntdll.dll",
        "opengl32.dll",
        "d3d11.dll",
        "dxgi.dll",
        "dwrite.dll",
        "comdlg32.dll",
        "comctl32.dll",
        "netapi32.dll",
        "userenv.dll",
        "crypt32.dll",
        "bcrypt.dll",
        "secur32.dll",
        "iphlpapi.dll",
        "mpr.dll",
        "dnsapi.dll",
        "authz.dll",
        "shlwapi.dll",
        "shcore.dll",
        "msvcrt.dll",
        "ucrtbase.dll",
        "nsi.dll",
        "win32u.dll",
        "rpcrt4.dll",
        "sechost.dll",
        "combase.dll",
        "setupapi.dll",
        "wtsapi32.dll",
        "d3d12.dll",
        "dbghelp.dll",
    ]
}


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


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("dest", type=Path, help="deploy directory containing exe/dlls")
    ap.add_argument(
        "--from",
        dest="search",
        type=Path,
        default=Path("/ucrt64/bin"),
        help="directory to search for missing DLLs",
    )
    args = ap.parse_args()
    dest: Path = args.dest
    search: Path = args.search
    if not dest.is_dir():
        print(f"ERROR: dest is not a directory: {dest}", file=sys.stderr)
        return 1
    if not search.is_dir():
        print(f"WARNING: search dir missing: {search}", file=sys.stderr)
        return 0

    index = {p.name.lower(): p for p in search.glob("*.dll")}
    changed, rounds = True, 0
    while changed and rounds < 20:
        changed = False
        rounds += 1
        for pe in list(dest.rglob("*.dll")) + list(dest.rglob("*.exe")):
            for name in imports(pe):
                low = name.lower()
                if low in SYS_DLL:
                    continue
                if any((dest / n).exists() for n in (name, name.lower(), Path(name).name)):
                    continue
                src = index.get(low)
                if src is None:
                    continue
                print(f"copy missing dep {src.name}")
                (dest / src.name).write_bytes(src.read_bytes())
                changed = True
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
