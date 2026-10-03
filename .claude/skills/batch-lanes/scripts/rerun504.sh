#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
repo=$(gh repo view --json nameWithOwner -q .nameWithOwner)
# rerun504.sh — for open batch PRs with a failed check whose job log shows an infra HTTP 5xx, rerun the failed jobs once the run has completed.
for pr in $(gh pr list --state open --author @me --limit 200 --json number -q '.[].number'); do
  for link in $(gh pr checks $pr --json bucket,link -q '.[]|select(.bucket=="fail")|.link' 2>/dev/null); do
    job=$(grep -oE '[0-9]+$' <<<"$link"); run=$(grep -oE 'runs/[0-9]+' <<<"$link" | cut -d/ -f2)
    gh api repos/$repo/actions/jobs/$job/logs 2>/dev/null | grep -qE "Unexpected HTTP response: 5[0-9][0-9]" || continue
    st=$(gh run view $run -R "$repo" --json status -q .status 2>/dev/null)
    if [ "$st" = completed ]; then gh run rerun $run --failed -R "$repo" >/dev/null 2>&1 && echo "RERUN $pr run $run (infra 5xx)"; else echo "WAIT-RERUN $pr run $run ($st)"; fi
  done
done
