#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LEDGER="$ROOT/.claude/skills/flake-ledger/scripts/ledger.py"
fail() { echo "FAIL: $*" >&2; exit 1; }
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# Existing JSONL rows without a signature remain readable.
printf '%s\n' '{"at":"2026-01-01T00:00:00Z","test":"LegacyTest","where":"pkg","run":"","branch":"main","note":"old row"}' > "$tmp/legacy.jsonl"
FLAKE_LEDGER="$tmp/legacy.jsonl" python3 "$LEDGER" check > "$tmp/legacy.out" || fail "legacy ledger row was not readable"
grep -q 'LegacyTest @ pkg' "$tmp/legacy.out" || fail "legacy row missing from check"

# Matching signatures across tests are reported by record and check.
export FLAKE_LEDGER="$tmp/shared.jsonl"
python3 "$LEDGER" record --test PageA --where duplicate-text --signature 'round end' > "$tmp/a.out"
python3 "$LEDGER" record --test PageB --where duplicate-text --signature 'round end' > "$tmp/b.out"
grep -q 'SHARED SIGNATURE: "round end" seen in 2 tests: PageA, PageB' "$tmp/b.out" || fail "record omitted shared signature"
python3 "$LEDGER" check > "$tmp/shared.out"
grep -q 'SHARED SIGNATURE' "$tmp/shared.out" || fail "check omitted shared signature"
grep -q 'PageA, PageB' "$tmp/shared.out" || fail "check omitted test list"

# A filtered check only reports shared signatures that include the selected test.
python3 "$LEDGER" record --test PageC --where different-text --signature 'another failure' > "$tmp/c.out"
python3 "$LEDGER" check --test PageC > "$tmp/page-c.out"
if grep -q 'SHARED SIGNATURE' "$tmp/page-c.out"; then fail "filtered check showed an unrelated shared signature"; fi
python3 "$LEDGER" check --test PageA > "$tmp/page-a.out"
grep -q 'SHARED SIGNATURE' "$tmp/page-a.out" || fail "filtered check omitted the selected test's shared signature"

# Repeating a signature in only one test is not shared.
export FLAKE_LEDGER="$tmp/same-test.jsonl"
python3 "$LEDGER" record --test PageA --where duplicate-text --signature 'same' > "$tmp/same1.out"
python3 "$LEDGER" record --test PageA --where duplicate-text --signature 'same' > "$tmp/same2.out"
if grep -q 'SHARED SIGNATURE' "$tmp/same1.out" "$tmp/same2.out"; then fail "same test was treated as shared"; fi
python3 "$LEDGER" check > "$tmp/same-check.out"
if grep -q 'SHARED SIGNATURE' "$tmp/same-check.out"; then fail "check treated same test as shared"; fi

# Different signatures across tests are not shared.
export FLAKE_LEDGER="$tmp/different.jsonl"
python3 "$LEDGER" record --test PageA --where duplicate-text --signature 'first' > "$tmp/diff1.out"
python3 "$LEDGER" record --test PageB --where duplicate-text --signature 'second' > "$tmp/diff2.out"
if grep -q 'SHARED SIGNATURE' "$tmp/diff1.out" "$tmp/diff2.out"; then fail "different signatures were treated as shared"; fi
python3 "$LEDGER" check > "$tmp/different-check.out"
if grep -q 'SHARED SIGNATURE' "$tmp/different-check.out"; then fail "check treated different signatures as shared"; fi

# Empty signatures are omitted from the stored row.
export FLAKE_LEDGER="$tmp/empty-signature.jsonl"
python3 "$LEDGER" record --test EmptySignature --where pkg --signature '' > "$tmp/empty.out"
python3 -c 'import json,sys; row=json.loads(open(sys.argv[1], encoding="utf-8").readline()); assert "signature" not in row' "$FLAKE_LEDGER" || fail "empty signature key was stored"

echo "flake-ledger tests passed"
