#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# rc.sh <slot> — receipt into rc<slot>.txt (trimmed), then "RC_DONE"
"$B/receipt.sh" "$1" 2>&1 | grep -vE '^ (M|A|\?\?) |\|' > "$BATCH_STATE/rc$1.txt"
# A delegation once rewrote a test file and dropped 21 of 22 tests (#8924). Flag any test file
# that loses more than 20 lines and more than it gains.
git -C "$WT_ROOT/wt-ib$1" diff --numstat HEAD -- '*.test.ts' '*.test.tsx' '*_test.go' \
  | awk -v allow="${ALLOW_SHRINK:-}" '$2 > 20 && $2 > $1 && index(" " allow " ", " " $3 " ") == 0 { print "WARN TEST_SHRINK " $3 " (+" $1 " -" $2 ")" }' >> "$BATCH_STATE/rc$1.txt"
# toHaveTextContent('') matches ANY content, so it proves nothing (use toBeEmptyDOMElement).
git -C "$WT_ROOT/wt-ib$1" diff -U0 HEAD -- '*.test.ts' '*.test.tsx' | grep -nE "^\+.*toHaveTextContent\((''|\"\")\)" \
  | sed 's/^/WARN VACUOUS_EMPTY_TEXT /' >> "$BATCH_STATE/rc$1.txt"
git -C "$WT_ROOT/wt-ib$1" diff -U0 HEAD -- "*.ts" "*.tsx" | grep -nE "^\+.*biome-ignore" | sed "s/^/WARN NEW_BIOME_IGNORE /" >> "$BATCH_STATE/rc$1.txt"
(cd "$WT_ROOT/wt-ib$1"/frontend && timeout 120 bun run deadcode 2>&1 | grep -E "^(Unused|Unlisted|Unresolved)" -A6 | sed "s/^/WARN DEADCODE /") >> "$BATCH_STATE/rc$1.txt"
echo RC_DONE >> "$BATCH_STATE/rc$1.txt"; cat "$BATCH_STATE/rc$1.txt"
