#!/usr/bin/env bash
# Build one or more workers and report raw/gzip size against the 1 MB free-tier limit.
#
# Mirrors the Makefile's build_worker macro exactly (same flags, same wasm-opt
# pass) so the numbers match CI byte for byte -- verified on extra2, which
# measured 594139 raw / 232016 gzip both locally and in CI. `make` is not
# installed on every dev box, which is why this exists as a shell script.
#
# Usage: measure.sh [worker ...]     (default: all six)
set -euo pipefail

cd "$(dirname "$0")/../../../.."
export PATH="$HOME/sdk/go1.25.8/bin:$HOME/.local/opt/tinygo/bin:$PATH"
export GOTOOLCHAIN=local          # TinyGo 0.40.1 refuses a newer toolchain
LIMIT=1048576

for w in "${@:-casino classic solo extra extra2 extra3}"; do
  mkdir -p "workers/$w/build"
  go run github.com/syumai/workers/cmd/workers-assets-gen -mode=tinygo -o "workers/$w/build" >/dev/null
  tinygo build -tags "$w" -o "workers/$w/build/app.wasm" -target wasm \
    -stack-size=128KB -no-debug -opt=z "./cmd/workers/$w"
  wasm-opt --enable-bulk-memory --enable-nontrapping-float-to-int --enable-sign-ext \
    -Oz "workers/$w/build/app.wasm" -o "workers/$w/build/app.wasm"
  raw=$(stat -c%s "workers/$w/build/app.wasm")
  gz=$(gzip -c "workers/$w/build/app.wasm" | wc -c)
  printf '%-8s %9d raw %9d gzip  %5.1f%% of limit  headroom %6.1f KB\n' \
    "$w" "$raw" "$gz" "$(echo "scale=4;$gz*100/$LIMIT" | bc)" "$(echo "scale=2;($LIMIT-$gz)/1024" | bc)"
done
