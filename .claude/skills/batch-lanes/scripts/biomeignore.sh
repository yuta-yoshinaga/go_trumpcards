#!/bin/bash
# Read a unified diff and warn only when biome-ignore comments increase per file.
awk -v allow=" ${ALLOW_BIOME_IGNORE:-} " '
  function emit() {
    if (file != "" && added > removed && index(allow, " " file " ") == 0)
      printf "WARN NEW_BIOME_IGNORE %s (+%d -%d)\n", file, added, removed
  }
  # A deleted file has "+++ /dev/null"; it must still start a new section, or its
  # removed ignores would offset the additions of the file before it.
  /^\+\+\+ / { emit(); file = ($0 ~ /^\+\+\+ b\//) ? substr($0, 7) : ""; added = 0; removed = 0; next }
  /^\+[^+].*biome-ignore/ { added++; next }
  /^-[^-].*biome-ignore/ { removed++; next }
  END { emit() }
'
