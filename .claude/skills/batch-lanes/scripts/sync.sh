#!/usr/bin/env bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
set -euo pipefail

usage() { echo "usage: sync.sh <numeric slot> (existing wt-ib<slot> required)" >&2; exit 2; }
[[ $# -eq 1 && "$1" =~ ^[0-9]+$ ]] || usage
s="$1"
lane="$WT_ROOT/wt-ib$s"
[[ -d "$lane" && -e "$lane/.git" ]] || usage

if "$B/busy.sh" "$s" >/dev/null; then
  echo "REFUSED: a delegation is running in wt-ib$s"
  exit 1
fi
cd "$lane"
if [[ "$(git rev-list --count origin/develop..HEAD)" != 0 ]]; then
  echo "REFUSED: lane has commits; merge by hand"
  exit 1
fi

patch="$BATCH_STATE/sync$s.patch"
untracked_backup="$BATCH_STATE/sync$s.untracked"
rm -rf "$untracked_backup"
mkdir -p "$untracked_backup"
git diff --binary HEAD > "$patch"
while IFS= read -r -d '' path; do
  mkdir -p "$untracked_backup/$(dirname "$path")"
  cp -a -- "$path" "$untracked_backup/$path"
done < <(git ls-files -o --exclude-standard -z)
echo "backup: $patch"
orig=$(git rev-parse HEAD)
orig_ref=$(git symbolic-ref --short -q HEAD || git rev-parse --short HEAD)

# Fetch before staging or committing the lane's WIP so a fetch failure leaves
# the lane's HEAD and working tree untouched.
if ! git fetch -q origin develop; then
  exit 1
fi

changed=0
mapfile -d '' -t tracked_paths < <(git diff --name-only -z HEAD)
mapfile -d '' -t untracked_paths < <(git ls-files -o --exclude-standard -z)
if ((${#tracked_paths[@]} || ${#untracked_paths[@]})); then
  changed=1
  for path in "${tracked_paths[@]}"; do
    if [[ -e "$path" || -L "$path" ]]; then
      git add -- "$path"
    else
      git rm -q --cached -- "$path"
    fi
  done
  for path in "${untracked_paths[@]}"; do git add -- "$path"; done
  git commit -q --no-verify -m "sync wip"
fi

if git merge --no-edit origin/develop; then
  git reset -q --soft origin/develop
  git reset -q
  short=$(git rev-parse --short origin/develop)
  echo "SYNCED wt-ib$s onto $short"
  git status --porcelain
  echo "re-run rcg.sh $s before ship.sh"
else
  while IFS= read -r -d '' path; do echo "CONFLICT $path"; done < <(git diff --name-only --diff-filter=U -z)
  git merge --abort 2>/dev/null || true
  if ((changed)); then
    git reset -q --soft "$orig"
    git reset -q
  fi
  echo "lane kept on $orig_ref; to rebase by hand: git stash -u && git merge --ff-only origin/develop && git stash pop"
  exit 2
fi
