#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# treehash.sh <slot> — hash of the worktree's uncommitted state (tracked diff + untracked files)
cd "$WT_ROOT/wt-ib$1" || exit 1
{ git diff HEAD; git ls-files -o --exclude-standard -z | xargs -0 -r sha1sum; } | sha1sum | cut -d' ' -f1
