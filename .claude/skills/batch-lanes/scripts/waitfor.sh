#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# usage: waitfor.sh <max-seconds> <shell condition>  — real wait (sleep is shimmed here)
end=$(( $(date +%s) + $1 )); shift
until eval "$*"; do [ $(date +%s) -ge $end ] && { echo TIMEOUT; exit 1; }; timeout 15 tail -f /dev/null; done; echo READY
