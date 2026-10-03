#!/usr/bin/env python3
"""Group repeated premises into policy candidates."""
import argparse, json
from pathlib import Path

def main():
    ap=argparse.ArgumentParser(); ap.add_argument("proposals"); ap.add_argument("--min",type=int,default=3); ap.add_argument("--policy",required=True); ap.add_argument("--individual",required=True); a=ap.parse_args()
    obj=json.loads(Path(a.proposals).read_text()); props=obj.get("proposals",obj) if isinstance(obj,dict) else obj
    groups={}
    for p in props: groups.setdefault(p.get("premise_key", ""),[]).append(p)
    policy=[]; individual=[]
    for key,items in sorted(groups.items()):
        unique=sorted(set(p.get("game") for p in items))
        if key and len(unique)>=a.min: policy.append({"premise_key":key,"games":unique,"proposals":items})
        else: individual.extend(items)
    Path(a.policy).write_text(json.dumps(policy,ensure_ascii=False,indent=2)+"\n")
    Path(a.individual).write_text(json.dumps(individual,ensure_ascii=False,indent=2)+"\n")
    print(json.dumps(policy,ensure_ascii=False,indent=2))
if __name__=="__main__":main()
