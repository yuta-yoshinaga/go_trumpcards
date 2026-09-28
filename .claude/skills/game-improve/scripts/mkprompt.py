#!/usr/bin/env python3
"""Render grouped game-improvement generation prompts."""
import argparse, json
from pathlib import Path

FALSE = """- Keyboard shortcuts come from shared hooks / `useActionShortcuts`.
- `CardImage` already provides a default alt through `cardAlt`.
- The video-poker family renders through `VideoPokerGameContent`.
- Live regions are often always-mounted sr-only elements with a testid.
- Column numbers are 0-based by repository convention; never propose 1-based numbering.
- Purely decorative motion (screen shake and similar) is low value.
- In `sync_cpu_loop` games, never propose showing the CPU's turn.
- Do not treat a config value as a play statistic (Whitehead `drawCount`)."""

EVIDENCE_GUIDANCE = """## Current state and absence evidence
- `current_state` is a required array of `{{path, line, note}}` identifying where the current behavior being criticized lives.
- `absence_evidence` remains required. Its `pattern` means code that would exist if the proposed improvement were already implemented.
- Example claim: "hand buttons do not say which card is a trump".
  - `current_state`: `frontend/src/pages/TutePage.tsx:351`, "passes trumpIndices only for styling".
  - Good pattern: `a11y\\.trump|trump.*aria-label|aria-label.*[Tt]rump` — it would match if a trump announcement existed.
  - Bad pattern: `trumpIndices=\\{trumpIndices\\}` — it matches the current code and says nothing about absence.
- A pattern that matches the code quoted in `current_state` is wrong."""

def main():
    ap=argparse.ArgumentParser(); ap.add_argument("group_index",type=int); ap.add_argument("--games",required=True); ap.add_argument("--size",type=int,default=12); a=ap.parse_args()
    games=json.loads(Path(a.games).read_text()); batch=games[a.group_index*a.size:(a.group_index+1)*a.size]
    if not batch: raise SystemExit(1)
    sections=[]
    for g in batch:
        page=[g["page"]] if g.get("page") else []
        presenters=[p for p in (g.get("web_presenter"),g.get("cui_presenter"),g.get("interactor")) if p]
        groups=(("Page",page),("Presenters / interactor",presenters),("Hooks",g.get("hooks",[])),("Components",g.get("components",[])),("Locales",g.get("locales",[])))
        surface="\n".join(f"{label}:\n"+"\n".join(f"- `{p}`" for p in sorted(set(paths))) for label,paths in groups)
        past="\n".join(f"- 既出: {title}" for title in g.get("past",[])) or "- 既出: なし"
        cpu=("yes — CPU turns run inside the human's request; a \"CPU is acting\" state never reaches the screen" if g.get("sync_cpu_loop") else "no")
        sections.append(f"### {g['game']} = {g.get('ja') or '日本語名不明'}\nsync_cpu_loop: {cpu}\nSurface files (read all before claiming anything is missing):\n{surface}\nNested shared components (the `deep` set) are also searched by the absence check, though they are not listed.\nPast issues:\n{past}")
    print(f"""go_trumpcards リポジトリの読み取り専用調査。各ゲームにつき最も効果の大きい改善を1件、JSONで返す。各Surface filesをすべて確認してから不足を主張すること。

## False positives to avoid
{FALSE}

## Requirements
- Code-backed proposal with concrete paths/symbols; avoid generic testing/refactor suggestions.
- Do not duplicate closed issues; verify current code.
- A past issue marked (not planned) was rejected because its premise was wrong; do not propose that premise again.
- Require `absence_evidence`: at least one {{claim, pattern, paths}}. A script will rerun each PCRE regex (as `git grep -P`) across the entire game surface and return the proposal if anything matches. Make the pattern specific to the missing feature, not a broad keyword.
- `premise_key` is a short kebab-case name for the missing premise, shared across games with the same premise.
- `needs_cpu_turn_state` is true only when the requested UI needs to show an in-progress CPU turn.
- Include `files_read` with repository paths actually inspected.
{EVIDENCE_GUIDANCE}
- `title` Japanese with a leading category tag and final （game Japanese name）. `body` has 背景・現状 / 提案 / 受け入れ条件 / 対象ファイル sections.

## Output
Return {{"proposals":[{{"game":"...","title":"...","body":"...","premise_key":"...","needs_cpu_turn_state":false,"files_read":[],"current_state":[{{"path":"...","line":1,"note":"..."}}],"absence_evidence":[{{"claim":"...","pattern":"...","paths":[]}}]}}]}}.

## Games
{"\n\n".join(sections)}""")
if __name__=="__main__":main()
