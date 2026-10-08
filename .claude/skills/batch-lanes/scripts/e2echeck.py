#!/usr/bin/env python3
"""warn when a ja locale value removed/changed in the working diff still appears in frontend/e2e"""
import subprocess, sys, re, os
wt = sys.argv[1]
os.chdir(wt)
diff = subprocess.run(['git','diff','-U0','--','frontend/src/i18n/locales/ja/'],capture_output=True,text=True).stdout
# Only the specs of the games whose locale changed (all specs when common.json changed):
# a removed fragment like "ラウンド:" also occurs in unrelated games' specs.
games = set(re.findall(r'^\+\+\+ b/frontend/src/i18n/locales/ja/([a-z0-9]+)\.json', diff, re.M))
specs = ['frontend/e2e'] if 'common' in games else [f'frontend/e2e/{g}.spec.ts' for g in games if os.path.exists(f'frontend/e2e/{g}.spec.ts')]
removed = []
for line in diff.splitlines():
    if line.startswith('-') and not line.startswith('---'):
        m = re.match(r'-\s*"[^"]+":\s*"(.*)",?\s*$', line)
        if m:
            # literal fragments between placeholders, 3+ chars
            for frag in re.split(r'\{\{[^}]+\}\}', m.group(1)):
                frag = frag.strip()
                if len(frag) >= 3 or (len(frag) >= 2 and not frag.isascii()):
                    removed.append(frag)
added = diff  # crude: skip fragments still present in the new values
for frag in sorted(set(removed)):
    if ('+' in added) and re.search(r'^\+.*' + re.escape(frag), added, re.M):
        continue
    if not specs:
        continue
    hits = subprocess.run(['git','grep','-n','-F',frag,'--',*specs],capture_output=True,text=True).stdout.splitlines()
    hits = [hit for hit in hits if len(hit.split(':', 2)) == 3 and
            not hit.split(':', 2)[2].lstrip().startswith(('//', '*', '/*'))]
    if hits:
        print(f'E2ECHECK WARN: removed ja text "{frag}" still in e2e: ' + hits[0])
