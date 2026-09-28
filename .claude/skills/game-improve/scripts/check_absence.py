#!/usr/bin/env python3
"""Validate proposals and search each claimed absence over the game's code surface."""
import argparse, json, re, subprocess, sys
from pathlib import Path
from surface_map import ROOT, build_one

REQUIRED={"game","title","body","premise_key","needs_cpu_turn_state","files_read","current_state","absence_evidence"}

def line_count(path):
    return len(path.read_text(errors="replace").splitlines())

def main():
    ap=argparse.ArgumentParser(); ap.add_argument("proposals"); ap.add_argument("--games",required=True); ap.add_argument("--clean",required=True); ap.add_argument("--suspect",required=True); a=ap.parse_args()
    try:
        obj=json.loads(Path(a.proposals).read_text()); props=obj.get("proposals") if isinstance(obj,dict) else obj
        games=json.loads(Path(a.games).read_text())
        if not isinstance(props,list) or not isinstance(games,list): raise ValueError("proposals and games must be arrays")
    except Exception as e:
        print(f"bad input: {e}",file=sys.stderr); return 2
    gm={x["game"]:x for x in games}; clean=[]; suspect=[]
    for p in props:
        reasons=[]; hits=[]
        if not isinstance(p,dict):
            suspect.append({"proposal":p,"reasons":["invalid-proposal"]}); continue
        missing=sorted(REQUIRED-set(p))
        if missing: reasons.append({"reason":"missing-fields","fields":missing})
        g=gm.get(p.get("game"))
        if g is None:
            reasons.append({"reason":"unknown-game"}); g=build_one(p.get("game","")) if p.get("game") else {}
        evidence=p.get("absence_evidence")
        if not isinstance(evidence,list) or not evidence: reasons.append({"reason":"missing-absence-evidence"}); evidence=[]
        current_state=p.get("current_state")
        if not isinstance(current_state,list):
            reasons.append({"reason":"invalid-current-state"}); current_state=[]
        current_locations=[]
        for idx,state in enumerate(current_state):
            if not isinstance(state,dict) or not all(k in state for k in ("path","line","note")):
                reasons.append({"reason":"invalid-current-state","index":idx}); continue
            path,line=state["path"],state["line"]
            if not isinstance(path,str) or not (ROOT/path).is_file():
                reasons.append({"reason":"missing-current-state-path","index":idx,"path":path}); continue
            try: count=line_count(ROOT/path)
            except OSError:
                reasons.append({"reason":"missing-current-state-path","index":idx,"path":path}); continue
            if not isinstance(line,int) or isinstance(line,bool) or line < 1 or line > count:
                reasons.append({"reason":"invalid-current-state-line","index":idx,"path":path,"line":line,"line_count":count}); continue
            current_locations.append((path,line))
        surface=set()
        if g:
            surface.update(g.get("components",[])); surface.update(g.get("hooks",[])); surface.update(g.get("deep",[])); surface.update(g.get("locales",[]))
            surface.update(x for x in (g.get("page"),g.get("web_presenter"),g.get("cui_presenter"),g.get("interactor")) if x)
        for idx,ev in enumerate(evidence):
            if not isinstance(ev,dict) or not all(k in ev for k in ("claim","pattern","paths")):
                reasons.append({"reason":"invalid-evidence","index":idx}); continue
            paths=ev["paths"]
            if not isinstance(paths,list):
                reasons.append({"reason":"invalid-evidence-paths","index":idx}); continue
            missing_paths=[path for path in paths if not isinstance(path,str) or not (ROOT/path).is_file()]
            if missing_paths:
                reasons.append({"reason":"missing-evidence-path","index":idx,"paths":missing_paths}); continue
            try: re.compile(ev["pattern"])
            except (TypeError,re.error) as e:
                reasons.append({"reason":"invalid-regex","index":idx,"error":str(e)}); continue
            search_paths=sorted(surface|set(paths))
            cmd=["git","grep","-nP","-e",ev["pattern"],"--",*search_paths]
            run=subprocess.run(cmd,cwd=ROOT,text=True,capture_output=True)
            if run.returncode not in (0,1):
                reasons.append({"reason":"grep-error","index":idx,"error":run.stderr.strip()})
            elif run.returncode==0:
                matches=run.stdout.splitlines()
                hits.append({"index":idx,"claim":ev["claim"],"hits":matches[:10]})
                reasons.append({"reason":"pattern-found","index":idx})
                near_current=[]
                for match in matches:
                    path,sep,rest=match.partition(":")
                    line_text,sep2,_=rest.partition(":")
                    if not sep or not sep2: continue
                    try: hit_line=int(line_text)
                    except ValueError: continue
                    if any(path==state_path and abs(hit_line-state_line)<=3 for state_path,state_line in current_locations):
                        near_current.append(match)
                if near_current:
                    reasons.append({"reason":"pattern-matches-current-state","index":idx,"hits":near_current[:10]})
        if p.get("needs_cpu_turn_state") is True and g.get("sync_cpu_loop"):
            reasons.append({"reason":"sync-cpu-loop"})
        if hits: p=dict(p,absence_hits=hits)
        (suspect if reasons else clean).append({"proposal":p,"reasons":reasons} if reasons else p)
    Path(a.clean).write_text(json.dumps(clean,ensure_ascii=False,indent=2)+"\n")
    Path(a.suspect).write_text(json.dumps(suspect,ensure_ascii=False,indent=2)+"\n")
    counts={}
    for item in suspect:
        for r in item.get("reasons",[]): counts[r["reason"]]=counts.get(r["reason"],0)+1
    print(f"clean={len(clean)} suspect={len(suspect)} reasons={json.dumps(counts,sort_keys=True)}")
    return 2 if any(any(r.get("reason","").startswith(("missing-", "invalid-", "unknown-")) for r in x.get("reasons",[])) for x in suspect) else 0
if __name__=="__main__": raise SystemExit(main())
