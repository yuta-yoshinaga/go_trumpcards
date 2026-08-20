#!/usr/bin/env python3
"""Move games between Cloudflare Worker size buckets.

Usage:
    move-game.py <target> <game> [<game> ...]
    move-game.py --check                       # report the current bucket layout

The bucket is a binary-size bucket, not a taxonomy (ADR-0027 / ADR-0032 /
ADR-0036). Moving a game changes nothing users can see -- only which TinyGo
WASM binary it compiles into.

Why a script: a move touches five registration points plus every production
file's build tag, and a miss is silent in four of the six places. Doing 35 of
them by hand is how a broken worker route ships. `go build` catches a missed
build tag loudly, but nothing catches a missed gameExec.ts entry except a user
hitting a 404, so the edits are made together or not at all.

Not handled: docs/cloudflare-workers.md's per-worker prose lists. They are free
text inside a table cell, not a machine-readable list, so the script reports
what to edit rather than guessing. See --check.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
BUCKETS = ["casino", "classic", "solo", "extra", "extra2", "extra3", "extra4"]
SRC_DIRS = ["internal/domain", "internal/usecase", "internal/adapter"]
REGISTRY = ROOT / "internal/infrastructure/games/registry.go"
REGISTRY_TEST = ROOT / "internal/infrastructure/games/registry_test.go"
GAME_EXEC = ROOT / "frontend/src/api/gameExec.ts"


def bucket_file(b: str) -> Path:
    return ROOT / f"internal/infrastructure/games/{b}/{b}.go"


class Tree:
    """Buffered file edits, flushed only once every game in the batch succeeds.

    Without this a failure partway through leaves the tree half-moved: the
    registration removed from one bucket but not added to the other, some build
    tags rewritten and some not. That state builds and looks plausible, which is
    worse than a crash. Nothing reaches disk until commit().
    """

    def __init__(self) -> None:
        self.buf: dict[Path, str] = {}

    def read(self, p: Path) -> str:
        if p not in self.buf:
            self.buf[p] = p.read_text(encoding="utf-8")
        return self.buf[p]

    def write(self, p: Path, text: str) -> None:
        self.buf[p] = text

    def commit(self) -> None:
        for p, text in self.buf.items():
            p.write_text(text, encoding="utf-8")


SCAFFOLD = '''
import (
\t"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
\t"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
\t"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
\t"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
\t"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
}
'''


def ensure_register_func(tree: Tree, bucket: str) -> None:
    """Give an empty Phase-1 bucket package the imports and init() to register into."""
    p = bucket_file(bucket)
    text = tree.read(p)
    if "func init()" in text:
        return
    text = text.replace(
        "// Currently empty. Phase 1 of ADR-0036 adds the bucket and proves the build and\n"
        "// deploy path; Phase 2 moves games in.\n//\n",
        "",
    ).replace(
        "// Currently empty. Phase 1 of ADR-0036 adds the bucket and proves the build and\n"
        "// deploy path; Phase 2 moves games in.\n",
        "",
    )
    tree.write(p, text.rstrip("\n") + "\n" + SCAFFOLD)


def const(b: str) -> str:
    """casino -> Casino, extra2 -> Extra2 (the Go/TS constant suffix)."""
    return b[0].upper() + b[1:]


def load_games() -> tuple[dict[str, str], dict[str, str]]:
    """Return (game -> Go type, game -> bucket), read from the RegisterKVGame calls."""
    pat = re.compile(
        r'RegisterKVGame\("([a-z0-9]+)",\s*games\.Category\w+,\s*\n\s*func\(\)\s+usecase\.(\w+)InteractorIF'
    )
    types: dict[str, str] = {}
    buckets: dict[str, str] = {}
    for b in BUCKETS:
        p = bucket_file(b)
        if not p.exists():
            continue
        for m in pat.finditer(p.read_text(encoding="utf-8")):
            types[m.group(1)] = m.group(2)
            buckets[m.group(1)] = b
    return types, buckets


def game_files(game_type: str, all_types: list[str]) -> list[Path]:
    """Production .go files owned by game_type.

    Ownership is longest-prefix: PokerSquares*.go belongs to PokerSquares, not
    Poker, so a file is only claimed when this type is the longest match. Test
    files are excluded -- they carry `//go:build test` and no category, so a
    move never touches them.
    """
    by_len = sorted(all_types, key=len, reverse=True)
    out = []
    for d in SRC_DIRS:
        for p in (ROOT / d).rglob("*.go"):
            if p.name.endswith("_test.go"):
                continue
            stem = p.name[:-3]
            low = stem.lower()
            match = next((t for t in by_len if low.startswith(t.lower())), None)
            if match == game_type:
                out.append(p)
    return sorted(out)


def retag(tree: Tree, path: Path, old: str, target: str) -> bool:
    """Rewrite `//go:build !js || !wasm || <old>` to name <target> instead.

    Returns False for files that do not carry the old bucket's tag -- untagged
    shared helpers (hand_eval.go and friends) compile into every worker and are
    dropped by TinyGo's dead-code elimination when unreferenced, so they must be
    left alone. A helper tagged for a single bucket that the moved game still
    needs will fail the build, which is the intended loud failure.
    """
    text = tree.read(path)
    first = text.split("\n", 1)[0]
    if not first.startswith("//go:build") or not re.search(rf"\b{old}\b", first):
        return False
    tree.write(path, text.replace(first, re.sub(rf"\b{old}\b", target, first), 1))
    return True


def sub_once(tree: Tree, path: Path, old: str, new: str) -> None:
    text = tree.read(path)
    if text.count(old) != 1:
        raise SystemExit(f"{path}: expected exactly 1 occurrence of {old!r}, found {text.count(old)}")
    tree.write(path, text.replace(old, new, 1))


def move_kv_block(tree: Tree, game: str, old: str, target: str) -> None:
    """Move the whole RegisterKVGame(...) call between the two bucket files."""
    src, dst = bucket_file(old), bucket_file(target)
    text = tree.read(src)
    start = text.index(f'\tgames.RegisterKVGame("{game}", games.Category{const(old)},')

    # Walk to the matching close paren of the call, so nested func literals are
    # kept intact. Counting parens is enough here: the registration blocks
    # contain no parens inside string or rune literals.
    depth, i = 0, text.index("(", start)
    while True:
        if text[i] == "(":
            depth += 1
        elif text[i] == ")":
            depth -= 1
            if depth == 0:
                break
        i += 1
    end = text.index("\n", i) + 1
    block = text[start:end].replace(f"games.Category{const(old)}", f"games.Category{const(target)}", 1)
    tree.write(src, text[:start] + text[end:])

    ensure_register_func(tree, target)
    dtext = tree.read(dst)
    close = dtext.rindex("}")  # closing brace of the destination's init()
    tree.write(dst, dtext[:close] + block + dtext[close:])


def bump_counts(tree: Tree, old: str, target: str) -> None:
    text = tree.read(REGISTRY_TEST)
    for bucket, delta in ((old, -1), (target, +1)):
        m = re.search(rf"(expected{const(bucket)}\s*=\s*)(\d+)", text)
        if not m:
            raise SystemExit(f"registry_test.go: no expected{const(bucket)} constant")
        text = text[: m.start(2)] + str(int(m.group(2)) + delta) + text[m.end(2) :]
    tree.write(REGISTRY_TEST, text)


def main() -> None:
    args = sys.argv[1:]
    types, buckets = load_games()

    if args[:1] == ["--check"]:
        for b in BUCKETS:
            n = sum(1 for g in buckets.values() if g == b)
            print(f"{b:8s} {n:3d} games")
        return

    if len(args) < 2 or args[0] not in BUCKETS:
        raise SystemExit(__doc__)
    target, names = args[0], args[1:]

    all_types = list(types.values())

    # Some games are variants implemented on another game's domain type: razz
    # and sevencardstud both use SevenCardStud, spanish21 rides on BlackJack,
    # irishpoker on Pineapple. Those share every production file, so a bucket
    # is a property of the implementation, not of the game. Moving one sibling
    # rewrites the other's build tags too and leaves it registered in a worker
    # whose binary no longer compiles -- caught here rather than 200 files later.
    siblings: dict[str, list[str]] = {}
    for g, t in types.items():
        siblings.setdefault(t, []).append(g)
    batch = set(names)
    for game in names:
        group = sorted(siblings[types[game]])
        if not set(group) <= batch:
            raise SystemExit(
                f"{game} shares the {types[game]} implementation with "
                f"{', '.join(g for g in group if g != game)}; move the whole group "
                f"or none of it:\n    move-game.py {target} {' '.join(group)}"
            )

    tree = Tree()
    for game in names:
        if game not in buckets:
            raise SystemExit(f"unknown game {game!r}")
        old = buckets[game]
        if old == target:
            raise SystemExit(f"{game} is already in {target}")

        moved = [p for p in game_files(types[game], all_types) if retag(tree, p, old, target)]
        sub_once(tree, REGISTRY, f'{{Name: "{game}", Category: Category{const(old)}}}',
                 f'{{Name: "{game}", Category: Category{const(target)}}}')
        move_kv_block(tree, game, old, target)
        bump_counts(tree, old, target)
        sub_once(tree, GAME_EXEC, f"  {game}: WORKER_{old.upper()},", f"  {game}: WORKER_{target.upper()},")
        print(f"{game}: {old} -> {target} ({len(moved)} files retagged)")

    tree.commit()
    print("\nStill to do by hand: docs/cloudflare-workers.md per-worker lists.")
    print("Then verify: go build ./... && go test -tags test ./internal/infrastructure/games/...")


if __name__ == "__main__":
    main()
