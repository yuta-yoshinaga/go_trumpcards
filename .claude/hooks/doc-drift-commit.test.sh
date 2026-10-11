#!/usr/bin/env bash
set -uo pipefail

HOOK="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/doc-drift-commit.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
cd "$TMP" || exit 1
git init -q .
git config user.email t@example.com
git config user.name test

fail=0
run() {
  local expect=$1 desc=$2 cmd=$3 out blocked=pass
  out=$(printf '{"tool_input":{"command":%s}}' "$(printf '%s' "$cmd" | jq -Rs .)" | bash "$HOOK")
  printf '%s' "$out" | jq -e . >/dev/null 2>&1 || { echo "  FAIL  $desc: invalid JSON: $out"; fail=1; return; }
  printf '%s' "$out" | grep -q '"permissionDecision":"deny"' && blocked=block
  if [ "$blocked" = "$expect" ]; then echo "  ok    $desc ($expect)"; else echo "  FAIL  $desc: expected $expect, got $blocked -- $out"; fail=1; fi
}
clear_index() { git read-tree --empty; }

echo 'doc-drift-commit:'
run pass 'non-commit command is ignored' 'git status'

mkdir -p docs/adr
printf '# ADR\n' > docs/adr/0100-test.md
git add docs/adr/0100-test.md
run block 'new ADR requires index' 'git commit -m x'
printf '# ADR index\n' > docs/adr/README.md
git add docs/adr/README.md
run pass 'new ADR with index is allowed' 'git commit -m x'
clear_index

mkdir -p internal/infrastructure/web
printf 'mux.HandleFunc("POST /new", handler)\n' > internal/infrastructure/web/routes.go
git add internal/infrastructure/web/routes.go
run block 'new registered route requires OpenAPI' 'git commit -m x'
mkdir -p api
printf 'openapi: 3.0.0\n' > api/openapi.yaml
git add api/openapi.yaml
run pass 'new route with OpenAPI is allowed' 'git commit -m x'
clear_index

mkdir -p internal/infrastructure/games docs/manual
printf 'var registry = []*Game{\n\t{Name: "newgame", Category: CategorySolo},\n}\n' > internal/infrastructure/games/registry.go
git add internal/infrastructure/games/registry.go
run block 'new game reports each missing document' 'git commit -m x'
printf 'new game\n' > README.md
printf 'new game\n' > docs/games.md
printf 'manual\n' > docs/manual/newgame.md
git add README.md docs/games.md docs/manual/newgame.md
run pass 'new game with all documents is allowed' 'git commit -m x'
clear_index

# A changed existing entry has no added registry literal.
printf 'var registry = []*Game{\n\t{Name: "existing", Category: CategorySolo},\n}\n' > internal/infrastructure/games/registry.go
git add internal/infrastructure/games/registry.go
git commit -qm base
sed -i 's/CategorySolo/CategoryClassic/' internal/infrastructure/games/registry.go
git add internal/infrastructure/games/registry.go
run pass 'existing game entry edit does not require docs' 'git commit -m x'

exit "$fail"
