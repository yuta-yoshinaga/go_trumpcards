#!/usr/bin/env bash
# Side-effect-free allowlist validation shared by batch-lanes scripts.
validate_allowlist() {
  local name="$1" raw="$2" pattern="$3" item had_noglob=0
  local -a items=()
  local invalid=0
  [[ $- == *f* ]] && had_noglob=1
  set -f
  read -r -a items <<<"$raw"
  ((had_noglob)) || set +f
  for item in "${items[@]}"; do
    if [[ ! "$item" =~ $pattern ]]; then
      printf 'WARN %s_INVALID %s\n' "$name" "$item"
      invalid=1
    fi
  done
  return "$invalid"
}
