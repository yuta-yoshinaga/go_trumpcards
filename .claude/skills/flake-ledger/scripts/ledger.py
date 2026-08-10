#!/usr/bin/env python3
"""Append-only ledger of test failures, so "flaky" becomes a count instead of a feeling.

A test that failed once and passed on re-run is not evidence of a flake -- it is one
observation with no comparison. The ledger keeps every sighting with where it happened, so
the second sighting at the same location can be recognised as the second, which is the point
at which a re-run stops being an explanation.

Ledger lives at .claude/.flake-ledger.jsonl (gitignored, local knowledge). Confirmed flakes
belong in a GitHub issue, not in this file.

Usage:
  ledger.py record --test <id> --where <loc> [--run <url>] [--branch <b>] [--note <text>]
  ledger.py check  [--test <id>] [--min <n>]
  ledger.py list   [--since <YYYY-MM-DD>]
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from collections import defaultdict
from datetime import datetime, timezone
from pathlib import Path

REPO = Path(__file__).resolve().parents[4]
LEDGER = REPO / ".claude" / ".flake-ledger.jsonl"

# Two sightings at the same location is the threshold at which a re-run stops being an
# explanation. One is an observation; the ledger exists so the second can be seen as second.
CONFIRM_AT = 2


def _now() -> str:
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def _branch() -> str:
    try:
        out = subprocess.run(
            ["git", "rev-parse", "--abbrev-ref", "HEAD"],
            cwd=REPO, capture_output=True, text=True, check=True,
        )
        return out.stdout.strip()
    except (subprocess.CalledProcessError, OSError):
        return "unknown"


def _read() -> list[dict]:
    if not LEDGER.exists():
        return []
    rows = []
    for line in LEDGER.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            rows.append(json.loads(line))
        except json.JSONDecodeError:
            # A corrupt line must not swallow the rest of the history, but it must also not
            # pass unnoticed -- a ledger that silently drops rows undercounts, which is the
            # one failure mode that matters here.
            print(f"warning: skipping unparseable ledger line: {line[:80]}", file=sys.stderr)
    return rows


def cmd_record(args: argparse.Namespace) -> int:
    LEDGER.parent.mkdir(parents=True, exist_ok=True)
    row = {
        "at": _now(),
        "test": args.test,
        "where": args.where,
        "run": args.run or "",
        "branch": args.branch or _branch(),
        "note": args.note or "",
    }
    with LEDGER.open("a", encoding="utf-8") as fh:
        fh.write(json.dumps(row, ensure_ascii=False) + "\n")

    same = [r for r in _read() if r.get("test") == args.test and r.get("where") == args.where]
    n = len(same)
    verdict = "CONFIRMED FLAKE" if n >= CONFIRM_AT else "UNCONFIRMED (1 sighting)"
    print(f"recorded: {args.test} @ {args.where}")
    print(f"sightings at this location: {n} -> {verdict}")
    if n < CONFIRM_AT:
        print("A single failure is not a flake. Investigate it as a real failure until a")
        print("second sighting at the same location says otherwise.")
    else:
        print(f"Seen {n}x. Open or update a GitHub issue; do not just re-run.")
        for r in same:
            print(f"  {r['at']}  {r.get('branch', '?')}  {r.get('run') or '(local)'}")
    return 0


def cmd_check(args: argparse.Namespace) -> int:
    rows = _read()
    if not rows:
        print("ledger is empty -- no failures recorded yet.")
        return 0

    groups: dict[tuple[str, str], list[dict]] = defaultdict(list)
    for r in rows:
        groups[(r.get("test", "?"), r.get("where", "?"))].append(r)

    if args.test:
        groups = {k: v for k, v in groups.items() if k[0] == args.test}
        if not groups:
            print(f"no sightings recorded for {args.test}")
            return 0

    threshold = args.min if args.min is not None else CONFIRM_AT
    confirmed = {k: v for k, v in groups.items() if len(v) >= threshold}
    single = {k: v for k, v in groups.items() if len(v) < threshold}

    print(f"ledger: {len(rows)} sightings across {len(groups)} test/location pairs\n")
    if confirmed:
        print(f"CONFIRMED (>= {threshold} sightings at the same location):")
        for (test, where), v in sorted(confirmed.items(), key=lambda kv: -len(kv[1])):
            runs = ", ".join(r.get("run") or "local" for r in v[-3:])
            print(f"  {len(v):>2}x  {test} @ {where}")
            print(f"       latest: {runs}")
    if single:
        print(f"\nUNCONFIRMED (< {threshold}) -- treat as real failures:")
        for (test, where), v in sorted(single.items()):
            print(f"  {len(v):>2}x  {test} @ {where}")
    return 0


def cmd_list(args: argparse.Namespace) -> int:
    rows = _read()
    if args.since:
        rows = [r for r in rows if r.get("at", "") >= args.since]
    for r in rows:
        print(f"{r.get('at')}  {r.get('branch','?'):<24} {r.get('test')} @ {r.get('where')}"
              f"  {r.get('run') or '(local)'}  {r.get('note','')}")
    print(f"\n{len(rows)} row(s). Ledger: {LEDGER.relative_to(REPO)}")
    return 0


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = p.add_subparsers(dest="cmd", required=True)

    rec = sub.add_parser("record", help="record one failure sighting")
    rec.add_argument("--test", required=True, help="test identifier, e.g. TestKaiserRedeal or SpadesPage.test.tsx > bids")
    rec.add_argument("--where", required=True, help="package/file the failure happened in, e.g. internal/domain")
    rec.add_argument("--run", help="CI run URL, or omit for a local run")
    rec.add_argument("--branch", help="branch (defaults to the current one)")
    rec.add_argument("--note", help="one line on what the failure looked like")
    rec.set_defaults(func=cmd_record)

    chk = sub.add_parser("check", help="show which failures have repeated")
    chk.add_argument("--test", help="restrict to one test id")
    chk.add_argument("--min", type=int, help=f"sightings needed to confirm (default {CONFIRM_AT})")
    chk.set_defaults(func=cmd_check)

    lst = sub.add_parser("list", help="dump the raw ledger")
    lst.add_argument("--since", help="ISO date lower bound, e.g. 2026-08-01")
    lst.set_defaults(func=cmd_list)

    args = p.parse_args()
    return args.func(args)


if __name__ == "__main__":
    sys.exit(main())
