#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# usage: land.sh <pr> — gate (with expected SHA = remote branch tip) then squash-merge. Exit 4 means LANE-GATING; ALLOW_LANE_GATING=1 bypasses it. Caller must have read reviews.
cd "$REPO"
pr="$1"
if ! grep -qx "$pr" "$BATCH_STATE/reviewed.txt"; then
  echo "REFUSED: PR $pr is not in reviewed.txt -- read its reviews first" >&2
  exit 1
fi
if [[ "${ALLOW_LANE_GATING:-}" != 1 ]]; then
  pr_files=$(gh pr view "$pr" --json files -q '.files[].path')
  while IFS= read -r pidfile; do
    [[ -s "$pidfile" ]] || continue
    slot="${pidfile##*/rcg}"; slot="${slot%.pid}"
    rcg_live "$slot" "$(cat "$pidfile")" || continue
    wt="$WT_ROOT/wt-ib$slot"
    [[ -d "$wt" ]] || continue
    mapfile -d '' -t lane_files < <(collect "$wt")
    for file in "${lane_files[@]}"; do
      if grep -Fxq -- "$file" <<<"$pr_files"; then
        echo "LANE-GATING slot=$slot $file"
        exit 4
      fi
    done
  done < <(find "$BATCH_STATE" -maxdepth 1 -type f -name 'rcg*.pid' -print | sort -V)
fi
if [[ "${BATCH_SKIP_CODECOV:-}" == 1 ]]; then
  echo "SKIP: codecov/patch check skipped by BATCH_SKIP_CODECOV=1" >&2
else
  files=$(gh pr view "$pr" --json files -q '.files[].path')
  if grep -Eq '^(frontend/src/|internal/|cmd/|pkg/).*\.(go|ts|tsx)$' <<<"$files" \
      && grep -Ev '(_test\.go|\.test\.ts|\.test\.tsx)$' <<<"$files" \
        | grep -Eq '^(frontend/src/|internal/|cmd/|pkg/).*\.(go|ts|tsx)$'; then
    checks=$(gh pr checks "$pr" --json name,bucket)
    codecov=$(jq -c '[.[] | select(.name == "codecov/patch")] | .[0] // empty' <<<"$checks")
    if [[ -z "$codecov" ]]; then
      echo "WAIT: codecov/patch has not reported on PR $pr" >&2
      exit 3
    fi
    bucket=$(jq -r '.bucket' <<<"$codecov")
    if [[ "$bucket" == pending ]]; then
      echo "WAIT: codecov/patch is still pending on PR $pr" >&2
      exit 3
    fi
    if [[ "$bucket" != pass ]]; then
      echo "REFUSED: codecov/patch is $bucket" >&2
      exit 1
    fi
  fi
fi
br=$(gh pr view "$pr" --json headRefName -q .headRefName)
if [[ "${ALLOW_LANE_PENDING:-}" != 1 ]]; then
  while IFS= read -r wt; do
    [[ -n "$wt" ]] || continue
    [[ -d "$wt" ]] || continue
    wt_br=$(git -C "$wt" branch --show-current 2>/dev/null) || continue
    [[ "$wt_br" == "$br" ]] || continue
    reason=''
    rest=${wt#"$WT_ROOT"/wt-ib}
    if [[ "$rest" != "$wt" && "$rest" =~ ^[0-9]+$ ]]; then
      slot=$rest
      if "$B/busy.sh" "$slot" >/dev/null 2>&1; then
        reason="delegation is still running (slot $slot)"
      fi
    fi
    if [[ -z "$reason" ]] && [[ -n "$(git -C "$wt" status --porcelain)" ]]; then
      reason='uncommitted changes'
    fi
    if [[ -z "$reason" ]]; then
      git -C "$wt" fetch -q origin "$br" 2>/dev/null || true
      if ! pending=$(git -C "$wt" rev-list --count "origin/$br..HEAD" 2>/dev/null); then
        echo "REFUSED: $wt cannot compare with origin/$br" >&2
        exit 1
      fi
      if [[ "$pending" =~ ^[0-9]+$ ]] && (( pending > 0 )); then
        reason="$pending unpushed commit(s)"
      fi
    fi
    if [[ -n "$reason" ]]; then
      echo "REFUSED: $wt has $reason" >&2
      exit 1
    fi
  done < <(git worktree list --porcelain | sed -n 's/^worktree //p')
fi
git fetch -q origin "$br"; sha=$(git rev-parse "origin/$br")
out=$(bash "${MERGE_GATE:-.claude/skills/land-pr/scripts/merge-gate.sh}" "$pr" "$sha" 2>&1)
echo "$out" | grep -q 'VERDICT: gate passed' || { echo "$out" | grep -vE '^\s+ok'; exit 1; }
merge_rc=0
gh pr merge "$pr" --squash >/dev/null || merge_rc=$?
if [[ "$merge_rc" -ne 0 ]]; then
  st=$(gh pr view "$pr" --json state -q .state 2>/dev/null || true)
  if [[ "$st" != MERGED ]]; then
    echo "Failed to merge PR $pr" >&2
    exit "$merge_rc"
  fi
fi
st=$(gh pr view "$pr" --json state -q .state); echo "PR $pr $st"
iss=$(gh pr view "$pr" --json closingIssuesReferences -q '.closingIssuesReferences[].number')
while IFS= read -r i; do [[ -n "$i" ]] && echo "  issue $i $(gh issue view "$i" --json state -q .state)"; done <<< "$iss"
git branch -D "$br" >/dev/null 2>&1; true
git push origin --delete "$br" >/dev/null 2>&1 || true
