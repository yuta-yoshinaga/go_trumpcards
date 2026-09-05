#!/usr/bin/env python3
"""Evaluate movability of games in bulk.

Usage: movability.py [--all]

Output includes both the total file count for a game and the number of files
actually retagged. Files without build tags do not require retagging but still
contribute to the final binary size. As seen in PR #7104, moving games with 0
tagged files (like daifugo or pitch) can still yield significant byte savings
(e.g., 72KB) because removing their references from the bucket sub-package
allows TinyGo's Dead Code Elimination (DCE) to drop them. Therefore, a unit
showing "retag 0" is still a valid and impactful move candidate.

Note: This script ignores underscores to correctly match snake_case files
(e.g. interfaces/). However, move-game.py's game_files() uses strict prefix
matching and misses them. You must manually retag these files (see ADR-0037).
"""

from __future__ import annotations
import collections, importlib.util, re, signal, sys
from pathlib import Path

BUILD_RE = re.compile(r"^//go:build .*$", re.M)

# The documented usage pipes this into `head`, and the report is 200+ lines, so the
# reader closes the pipe every time. Without this, Python prints a BrokenPipeError
# traceback to stderr after a perfectly successful run.
signal.signal(signal.SIGPIPE, signal.SIG_DFL)

HERE = Path(__file__).resolve().parent

_spec = importlib.util.spec_from_file_location("mg", HERE / "move-game.py")
mg = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(mg)

_spec2 = importlib.util.spec_from_file_location("cr", HERE / "crossrefs.py")
cr = importlib.util.module_from_spec(_spec2)
_spec2.loader.exec_module(cr)

def find_wccs(nodes, edges):
    graph = collections.defaultdict(list)
    for u, v in edges:
        graph[u].append(v)
        graph[v].append(u)
    seen = set()
    wccs = []
    for node in nodes:
        if node not in seen:
            comp = set()
            q = [node]
            while q:
                curr = q.pop()
                if curr not in comp:
                    comp.add(curr)
                    seen.add(curr)
                    q.extend(graph[curr])
            wccs.append(comp)
    return wccs

def main() -> None:
    # Reject unknown flags the way move-game.py does. Silently ignoring them
    # means a typo like `--al` prints the default view and reads as success.
    unknown = [a for a in sys.argv[1:] if a != "--all"]
    if unknown:
        raise SystemExit(f"unknown argument(s): {' '.join(unknown)}\n{__doc__}")
    show_all = "--all" in sys.argv
    types, buckets_map = mg.load_games()
    all_types = list(types.values())
    by_len = sorted(list(set(all_types)), key=len, reverse=True)
    
    type_all_files = collections.defaultdict(list)
    file_buckets = collections.defaultdict(set)
    file_decls = {}
    file_refs = {}
    file_type = {}
    file_size = {}
    
    for d in mg.SRC_DIRS:
        for p in (mg.ROOT / d).rglob("*.go"):
            if p.name.endswith(("_test.go", "_mock.go")): continue
            s = p.read_text(encoding="utf-8")
            
            file_size[p] = len(s.encode("utf-8"))
            stem = p.name[:-3].lower()
            match = next((t for t in by_len if stem.replace("_", "").startswith(t.lower())), None)
            file_type[p] = match
            if match:
                type_all_files[match].append(p)
                
            # Search for the constraint rather than reading line 1: today all 4,161
            # tagged files put it first, but a file with a licence header above it
            # would read as untagged, and "untagged" is the direction that produces
            # a false CLEAN -- the expensive kind of wrong answer here.
            m = BUILD_RE.search(s)
            if not m: continue
            tags = set(re.findall(r"\b(casino|classic|solo|extra[2345]?)\b", m.group(0)))
            if not tags: continue
                
            file_buckets[p] = tags
            file_decls[p] = cr.package_decls(s) - {"_"}
            file_refs[p] = set(cr.REF.findall(s)) - {"_"}
            
    type_decls = collections.defaultdict(set)
    type_refs = collections.defaultdict(set)
    type_files = collections.defaultdict(list)
    shared_decls = collections.defaultdict(set)
    shared_refs = collections.defaultdict(set)
    
    for p, tags in file_buckets.items():
        t = file_type[p]
        if t:
            type_decls[t] |= file_decls[p]
            type_refs[t] |= file_refs[p]
            type_files[t].append(p)
        else:
            for b in tags:
                shared_decls[b] |= file_decls[p]
                shared_refs[b] |= file_refs[p]

    type_to_bucket = {t: buckets_map[g] for g, t in types.items()}
    base_units = []
    edges = set()
    
    for t in set(all_types):
        b = type_to_bucket[t]
        lose, lack = set(), set()
        
        for u in set(all_types):
            if t != u and type_to_bucket[u] == b:
                t_loses_to_u = type_decls[t] & type_refs[u]
                t_lacks_u = type_refs[t] & type_decls[u]
                if t_loses_to_u: lose |= t_loses_to_u
                if t_lacks_u: lack |= t_lacks_u
                if t_loses_to_u or t_lacks_u: edges.add((t, u))
                    
        lose |= (type_decls[t] & shared_refs[b])
        lack |= (type_refs[t] & shared_decls[b])
        
        games = [g for g, type_ in types.items() if type_ == t]
        size = sum(file_size[p] for p in type_all_files[t])
        
        base_units.append({
            'bucket': b,
            'types': {t},
            'games': games,
            'file_count': len(type_all_files[t]),
            'retag_count': len(type_files[t]),
            'bytes': size,
            'lose': len(lose),
            'lack': len(lack)
        })

    nodes = set(all_types)
    wccs = find_wccs(nodes, edges)
    
    welded_units = []
    for wcc in wccs:
        if len(wcc) <= 1: continue
        b = type_to_bucket[next(iter(wcc))]
        wcc_decls, wcc_refs = set(), set()
        wcc_files, wcc_retags, wcc_bytes, wcc_games = 0, 0, 0, []
        
        for t in wcc:
            wcc_decls |= type_decls[t]
            wcc_refs |= type_refs[t]
            wcc_files += len(type_all_files[t])
            wcc_retags += len(type_files[t])
            wcc_bytes += sum(file_size[p] for p in type_all_files[t])
            wcc_games.extend([g for g, type_ in types.items() if type_ == t])
            
        lose = wcc_decls & shared_refs[b]
        lack = wcc_refs & shared_decls[b]
        
        welded_units.append({
            'bucket': b,
            'types': wcc,
            'games': wcc_games,
            'file_count': wcc_files,
            'retag_count': wcc_retags,
            'bytes': wcc_bytes,
            'lose': len(lose),
            'lack': len(lack)
        })
        
    all_units = base_units + welded_units
    all_units.sort(key=lambda x: (x['bucket'], -x['bytes']))
    
    out_clean, out_all = [], []
    for u in all_units:
        is_clean = u['lose'] == 0 and u['lack'] == 0
        is_welded = len(u['types']) > 1
        games_str = ', '.join(sorted(u['games']))
        base = f"{u['bucket']:8s} / {games_str} / files: {u['file_count']:2d} (retag {u['retag_count']}) / {u['bytes']:7d} bytes"
        
        if is_clean:
            out_clean.append("[CLEAN] " + base)
            line = "[CLEAN] " + base
        else:
            line = base + f" (lose: {u['lose']:2d}, lack: {u['lack']:2d})"
            
        if not is_welded or is_clean or show_all:
            out_all.append(line)
            
    if show_all:
        for line in out_all: print(line)
    else:
        for line in out_clean: print(line)

if __name__ == "__main__":
    main()
