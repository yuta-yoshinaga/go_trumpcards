#!/usr/bin/env bash
# Side-effect-free allowlist validation shared by batch-lanes scripts.
validate_allowlist() {
  local name="$1" raw="$2" pattern="$3" item
  local invalid=0
  for item in $raw; do
    if [[ ! "$item" =~ $pattern ]]; then
      printf 'WARN %s_INVALID %s\n' "$name" "$item"
      invalid=1
    fi
  done
  return "$invalid"
}
