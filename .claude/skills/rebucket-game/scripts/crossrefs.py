#!/usr/bin/env python3
"""Report package-level symbols that would cross a bucket boundary after a move.

Usage: crossrefs.py <target> <game> [<game> ...]

Rebucketing is not just retagging: a game's files may declare package-level
symbols that games left behind still use, or use symbols declared in files that
stay. Go reports these as `undefined: X` -- but only after a 3.5-minute TinyGo
build, one symbol batch at a time. This does the same check statically in about
a second, in both directions, before anything is edited.

The classic example is GameResult: a core enum used across the whole casino
bucket that happens to be declared in BlackJack.go. Moving blackjack out
therefore breaks every casino game that reports a result -- nothing about the
game's own code hints at that.
"""

from __future__ import annotations

import collections
import importlib.util
import re
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
_spec = importlib.util.spec_from_file_location("mg", HERE / "move-game.py")
mg = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(mg)

KEYWORDS = {
    "break", "case", "chan", "const", "continue", "default", "defer", "else",
    "fallthrough", "for", "func", "go", "goto", "if", "import", "interface",
    "map", "package", "range", "return", "select", "struct", "switch", "type", "var",
}
FUNC = re.compile(r"^func\s+([A-Za-z_]\w*)\s*[\(\[]")
TYPE = re.compile(r"^type\s+([A-Za-z_]\w*)")
SINGLE = re.compile(r"^(?:var|const)\s+([A-Za-z_]\w*)")
OPEN = re.compile(r"^(?:var|const)\s*\($")
INBLOCK = re.compile(r"^\t([A-Za-z_]\w*)")
REF = re.compile(r"(?<![.\w])([A-Za-z_]\w*)")


def package_decls(text: str) -> set[str]:
    """Package-level declarations only: no methods, no locals.

    Methods are excluded deliberately -- `func (h *Holdem) Reset()` is reached
    through its receiver's type, so it never becomes an undefined package-level
    symbol when files move. Counting them makes every game look connected to
    every other via common method names like Reset or Output.
    """
    out: set[str] = set()
    inblock = False
    for line in text.split("\n"):
        if inblock:
            if line.startswith(")"):
                inblock = False
            else:
                m = INBLOCK.match(line)
                if m and m.group(1) not in KEYWORDS:
                    out.add(m.group(1))
            continue
        if OPEN.match(line):
            inblock = True
            continue
        for pat in (FUNC, TYPE, SINGLE):
            m = pat.match(line)
            if m:
                out.add(m.group(1))
                break
    return out - KEYWORDS


def main() -> None:
    if len(sys.argv) < 3:
        raise SystemExit(__doc__)
    target, names = sys.argv[1], sys.argv[2:]
    types, buckets = mg.load_games()
    all_types = list(types.values())
    moving_types = {types[g] for g in names}
    by_type = collections.defaultdict(list)
    for g, t in types.items():
        by_type[t].append(g)

    decls: dict[str, set[str]] = {}
    refs: dict[str, set[str]] = {}
    for t in by_type:
        d, r = set(), set()
        for p in mg.game_files(t, all_types):
            s = p.read_text(encoding="utf-8")
            d |= package_decls(s)
            r |= set(REF.findall(s))
        decls[t], refs[t] = d, r

    src_buckets = {buckets[by_type[t][0]] for t in moving_types}
    stay = [t for t in by_type if buckets[by_type[t][0]] in src_buckets and t not in moving_types]
    move = sorted(moving_types)

    broken_old: dict[str, list[str]] = collections.defaultdict(list)
    for t in move:
        for u in stay:
            for sym in decls[t] & refs[u]:
                broken_old[sym].append(u)
    broken_new: dict[str, list[str]] = collections.defaultdict(list)
    for u in stay:
        for t in move:
            for sym in decls[u] & refs[t]:
                broken_new[sym].append(t)

    print(f"moving {len(names)} games ({len(move)} units) out of {', '.join(sorted(src_buckets))} -> {target}\n")
    print(f"[1] symbols the OLD bucket would lose ({len(broken_old)}):")
    for sym, us in sorted(broken_old.items(), key=lambda kv: -len(kv[1]))[:25]:
        print(f"    {sym:32s} used by {len(us):2d} staying units e.g. {', '.join(sorted(us)[:3])}")
    print(f"\n[2] symbols the NEW bucket would lack ({len(broken_new)}):")
    for sym, ts in sorted(broken_new.items(), key=lambda kv: -len(kv[1]))[:25]:
        print(f"    {sym:32s} needed by {len(ts):2d} moving units e.g. {', '.join(sorted(ts)[:3])}")
    # Shared helper files belong to no game, so the unit-to-unit comparison above
    # cannot see them. They are the other way a move fails: betting.go is tagged
    # casino-only and solitaire_output_helper.go was `solo || extra`, so a game
    # moving elsewhere loses them. Reported separately because the fix differs --
    # widen the helper's tag rather than reconsider the move.
    owned = {p.resolve() for t in by_type for p in mg.game_files(t, all_types)}
    tagre = re.compile(r"\b(casino|classic|solo|extra[23]?)\b")
    helpers = []
    for d in mg.SRC_DIRS:
        for p in (mg.ROOT / d).rglob("*.go"):
            if p.name.endswith("_test.go") or p.resolve() in owned:
                continue
            first = p.read_text(encoding="utf-8").split("\n", 1)[0]
            if not first.startswith("//go:build"):
                continue  # untagged: compiled into every worker, always available
            buckets_ok = set(tagre.findall(first))
            if buckets_ok and target not in buckets_ok:
                helpers.append((p, buckets_ok, package_decls(p.read_text(encoding="utf-8"))))

    missing: dict[str, set[str]] = collections.defaultdict(set)
    for p, ok, d in helpers:
        for t in move:
            used = d & refs[t]
            if used:
                missing[f"{p.relative_to(mg.ROOT)}  [{'|'.join(sorted(ok))}]"] |= used

    print(f"\n[3] shared helpers not tagged for {target} ({len(missing)}):")
    for f, syms in sorted(missing.items()):
        shown = ", ".join(sorted(syms)[:4])
        print(f"    {f}\n        needs || {target}  (uses {shown}{', ...' if len(syms) > 4 else ''})")

    if not broken_old and not broken_new and not missing:
        print("\nclean: no package-level symbol crosses the boundary")


if __name__ == "__main__":
    main()
