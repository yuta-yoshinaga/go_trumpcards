#!/usr/bin/env python3
"""Clone an existing game's whole footprint into a new game name.

Usage: clone-game.py <src-key> <dst-key> <SrcType> <DstType> [--jp SrcJP DstJP] [--apply]

Cloning a near-identical game is the dominant pattern in this repo (Alaska from
RussianSolitaire, Fortress from BeleagueredCastle, Somerset from Fortress). Doing
it by hand costs an hour and reliably misses files, because the helper set differs
per game: RussianSolitaire has no hooks/use<X>Game.ts, Fortress has three utils
Alaska never needed, and MoveZone types are re-exported through api/games/mao.ts.

This finds the source's real footprint by grep instead of a fixed checklist, then
copies every file whose PATH carries the game key, applying the identifier
renames. It does NOT touch shared registration files (registry.go, gameApi.ts,
...) -- those need ordered insertion, so they stay manual.

Dry-run by default; pass --apply to write.

After running, three things still need doing by hand:

1. The 17-ish shared registration files it lists (ordered insertion).
2. descriptions.go and the shared locale files still need the Japanese display
   name by hand. Pass --jp to rename it inside the per-game files (it shows up in
   manuals, locale titles, Go doc comments AND page test assertions -- missing it
   fails the page test with "Unable to find text: /<oldname>/").
3. GENERIC helper names in the cloned _test.go. Anything without the game in its
   name (cardSpec, tableauFixture, ...) collides in the shared package_test and
   the build fails with "redeclared in this block". Prefix them.
"""
from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
SEARCH = ["internal", "frontend/src", "docs/manual"]


def variants(key: str, typ: str) -> list[tuple[str, str]]:
    """Identifier spellings a game uses, longest first so nested ones win."""
    lower, upper = key.lower(), key.upper()
    camel = typ[0].lower() + typ[1:]
    return [(typ, None), (camel, None), (upper, None), (lower, None)]


def rename(text: str, s_key, d_key, s_type, d_type) -> str:
    pairs = [
        (s_type, d_type),                                    # BeleagueredCastle -> Fortress
        (s_type[0].lower() + s_type[1:], d_type[0].lower() + d_type[1:]),
        (s_key.upper(), d_key.upper()),                      # FORTRESS_HELP
        (s_key, d_key),                                      # fortress
        (s_key.capitalize(), d_key.capitalize()),
    ]
    for old, new in pairs:
        if old and old != new:
            text = text.replace(old, new)
    return text


def main() -> None:
    argv = sys.argv[1:]
    apply = "--apply" in argv
    jp = None
    if "--jp" in argv:
        i = argv.index("--jp")
        jp = (argv[i + 1], argv[i + 2])
        argv = argv[:i] + argv[i + 3:]
    args = [a for a in argv if a != "--apply"]
    if len(args) != 4:
        raise SystemExit(__doc__)
    s_key, d_key, s_type, d_type = args

    out = subprocess.run(
        ["grep", "-rl", "-i", s_key, *SEARCH],
        cwd=ROOT, capture_output=True, text=True,
    ).stdout.split()
    # Content grep alone is not enough: a per-game file can go a whole file
    # without naming its own key. docs/manual/web/<game>.md is written entirely
    # in Japanese and never spells the ascii key, so cloning Fortress by content
    # silently skipped its Web manual. Union the content hits with a filename
    # walk so a file is found by EITHER route.
    by_name = [
        str(p.relative_to(ROOT))
        for d in SEARCH
        for p in (ROOT / d).rglob("*")
        if p.is_file() and s_key in p.name.lower()
    ]
    # only files whose NAME carries the key are per-game files; the rest are
    # shared registration points and need ordered manual insertion.
    per_game = sorted(set(by_name) | {f for f in out if s_key in Path(f).name.lower()})
    shared = sorted(set(out) - set(per_game))

    print(f"per-game files to clone ({len(per_game)}):")
    for f in sorted(per_game):
        src = ROOT / f
        dst = ROOT / f.replace(Path(f).name, rename(Path(f).name, s_key, d_key, s_type, d_type))
        print(f"  {f}\n    -> {dst.relative_to(ROOT)}")
        if apply:
            dst.parent.mkdir(parents=True, exist_ok=True)
            data = src.read_bytes()
            crlf = b"\r\n" in data
            text = rename(data.decode("utf-8"), s_key, d_key, s_type, d_type)
            if jp:
                text = text.replace(jp[0], jp[1])
            dst.write_bytes(text.encode("utf-8"))
            if crlf and b"\r\n" not in text.encode("utf-8"):
                raise SystemExit(f"CRLF lost writing {dst}")

    print(f"\nshared files needing MANUAL ordered insertion ({len(shared)}):")
    for f in shared:
        print(f"  {f}")
    if not apply:
        print("\n(dry run -- pass --apply to write)")


if __name__ == "__main__":
    main()
