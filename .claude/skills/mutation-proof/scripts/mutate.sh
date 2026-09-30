#!/usr/bin/env bash
# Temporarily replace one exact literal, run tests, restore, and run them again.
set -uo pipefail
usage() { echo "Usage: mutate.sh <file> <old-literal> <new-literal> -- <test command...>" >&2; exit 2; }
[[ $# -ge 5 ]] || usage
file=$1 old=$2 new=$3; shift 3
[[ ${1:-} == -- ]] || usage
shift
[[ $# -gt 0 && -f "$file" ]] || usage
backup=$(mktemp) || exit 2
cp -- "$file" "$backup" || { rm -f "$backup"; exit 2; }
restore() { cp -- "$backup" "$file"; rm -f -- "${backup}" "${backup}".*; }
trap restore EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
python3 - "$file" "$old" "$new" <<'PY'
from pathlib import Path
import sys
p=Path(sys.argv[1]); text=p.read_text(); old=sys.argv[2]
n=text.count(old)
if n != 1:
    print(f"mutate.sh: expected old literal exactly once, found {n}; aborting", file=sys.stderr)
    sys.exit(3)
p.write_text(text.replace(old,sys.argv[3],1))
PY
replace_rc=$?
if [[ $replace_rc -ne 0 ]]; then exit "$replace_rc"; fi
set +e
"$@" >"${backup}.mutated-output" 2>&1
mutated_rc=$?
cat "${backup}.mutated-output"
failures=$(python3 - "${backup}.mutated-output" <<'PY'
import re,sys
s=open(sys.argv[1],errors='replace').read()
nums=[int(x) for x in re.findall(r'Tests\s+(\d+)\s+failed',s)]
nums.extend([len(re.findall(r'^--- FAIL:',s,re.M))] if 'go test' in s else [])
print(max(nums,default=0))
PY
)
threshold=${MUTATION_MAX_FAILURES:-3}
if (( failures > threshold )); then
  echo "WARNING: mutation caused $failures failing tests (limit $threshold); it may have crashed the module instead of exercising the target. Try a narrower, behavior-preserving mutation." >&2
fi
restore
trap - EXIT INT TERM
"$@" >"${backup}.restored-output" 2>&1
restored_rc=$?
cat "${backup}.restored-output"
rm -f "${backup}.mutated-output" "${backup}.restored-output"
if (( restored_rc != 0 )); then
  echo "BROKEN BASELINE: restored code fails"
  exit 1
elif (( mutated_rc == 0 )); then
  echo "NOT PROVEN: tests still pass with the mutation (vacuous test or wrong anchor)"
  exit 1
else
  echo "PROVEN: mutation turned the test red and the restored code is green"
  exit 0
fi
