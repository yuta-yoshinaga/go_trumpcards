#!/usr/bin/env python3
"""Build game surface data and closed-issue history."""
import argparse, json, re, subprocess
from pathlib import Path
from surface_map import ROOT, build_one, registry_games

def main():
    ap = argparse.ArgumentParser(); ap.add_argument("--out", required=True); args = ap.parse_args()
    raw = subprocess.run(["gh", "issue", "list", "--state", "closed", "--limit", "2000", "--json", "number,title,stateReason"], cwd=ROOT, text=True, capture_output=True, check=True)
    issues = json.loads(raw.stdout)
    games = [build_one(g) for g in registry_games()]
    for game in games:
        ja = game["ja"] or ""
        matches = [i for i in issues if i.get("title", "").endswith(f"（{ja}）")]
        game["past"] = [(i["title"] + (" (not planned)" if i.get("stateReason") == "NOT_PLANNED" else "")) for i in matches[:8]]
    Path(args.out).write_text(json.dumps(games, ensure_ascii=False, indent=2) + "\n")
if __name__ == "__main__": main()
