#!/usr/bin/env python3
"""Render the verification prompt, including checker findings for the batch."""
import argparse, json
from pathlib import Path
from mkprompt import FALSE, EVIDENCE_GUIDANCE

def main():
    ap=argparse.ArgumentParser(); ap.add_argument("group_index",type=int); ap.add_argument("--games",required=True); ap.add_argument("--proposals",required=True); ap.add_argument("--hits"); ap.add_argument("--size",type=int,default=12); a=ap.parse_args()
    all_games=json.loads(Path(a.games).read_text()); games={x["game"]:x for x in all_games}
    props=json.loads(Path(a.proposals).read_text()); props=props.get("proposals",props) if isinstance(props,dict) else props
    props=[p.get("proposal",p) for p in props]
    props=props[a.group_index*a.size:(a.group_index+1)*a.size]
    findings=json.loads(Path(a.hits).read_text()) if a.hits else {"clean":props,"suspect":[]}
    suspects=findings if isinstance(findings,list) else findings.get("suspect",[]); chunks=[]
    for p in props:
        g=games.get(p.get("game"),{})
        page=[g["page"]] if g.get("page") else []
        presenters=[x for x in (g.get("web_presenter"),g.get("cui_presenter"),g.get("interactor")) if x]
        groups=(("Page",page),("Presenters / interactor",presenters),("Hooks",g.get("hooks",[])),("Components",g.get("components",[])),("Locales",g.get("locales",[])))
        surface="\n".join(f"{label}:\n"+"\n".join(f"- `{x}`" for x in sorted(set(paths))) for label,paths in groups)
        past="\n".join(f"- 既出: {title}" for title in g.get("past",[])) or "- 既出: なし"
        cpu=("yes — CPU turns run inside the human's request; a \"CPU is acting\" state never reaches the screen" if g.get("sync_cpu_loop") else "no")
        hit=next((x for x in suspects if x.get("proposal",x).get("game")==p.get("game")),None)
        chunks.append(f"### {p.get('game')}\nsync_cpu_loop: {cpu}\nSurface files (inspect all):\n{surface}\nNested shared components (the `deep` set) are also searched by the absence check, though they are not listed.\nPast issues:\n{past}\n\nProposal: {json.dumps(p,ensure_ascii=False,indent=2)}\n\nAbsence checker hits: {json.dumps(hit,ensure_ascii=False) if hit else 'none'}")
    print(f"""Read-only code review. Verify every factual claim against current code and every listed surface file. Absence patterns use PCRE regex (as `git grep -P`).
Known false positives:
{FALSE}
{EVIDENCE_GUIDANCE}
Keep only if evidence supports it and value is meaningful; fixed means correct factual errors; replaced means false, duplicated, low value, or already present. A past issue marked (not planned) was rejected because its premise was wrong; do not propose that premise again. Return all proposals with verdict and evidence (path:line). Maintain premise_key, needs_cpu_turn_state, files_read, absence_evidence in every output.

{chr(10).join(chunks)}""")
if __name__=="__main__":main()
