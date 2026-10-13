#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/png"
printf 'png' > "$tmp/png/foo-play.png"
printf 'png' > "$tmp/png/bar.png"
cat > "$tmp/bin/gh" <<'MOCK'
#!/usr/bin/env bash
printf 'gh %s\n' "$*" >> "$MOCK_LOG"
case "$1 $2" in
  'auth status') exit 0 ;;
  'api user') echo testuser ;;
  'repo view') echo testuser/project ;;
  'gist create') [[ ${MOCK_GIST_FAIL:-0} == 1 ]] && exit 1; echo https://gist.github.com/abc123 ;;
  'issue comment')
    shift 2
    while (($#)); do
      case $1 in --body-file) cp "$2" "$MOCK_BODY"; shift 2;; *) shift;; esac
    done
    echo https://github.com/testuser/project/issues/1#issuecomment-987
    ;;
  'api repos/testuser/project/issues/comments/987')
    if [[ ${MOCK_COUNT_MISMATCH:-0} == 1 ]]; then echo '<p><img src="https://gist.githubusercontent.com/testuser/abc123/raw/foo.png"></p>'
    elif [[ ${MOCK_NO_IMAGES:-0} == 1 ]]; then echo '<p>none</p>'
    elif [[ ${MOCK_CAMO:-0} == 1 ]]; then
      # GitHub proxies third-party hosts: src is camo, the original URL moves to data-canonical-src.
      echo '<p><img src="https://camo.githubusercontent.com/aa/foo" data-canonical-src="https://files.catbox.moe/foo.png" style="max-width: 100%;"><img src="https://camo.githubusercontent.com/bb/bar" data-canonical-src="https://files.catbox.moe/bar.png" style="max-width: 100%;"></p>'
    else
      echo '<p><img src="https://gist.githubusercontent.com/testuser/abc123/raw/foo-play.png"><img src="https://gist.githubusercontent.com/testuser/abc123/raw/bar.png"></p>'
    fi
    ;;
  *) exit 1 ;;
esac
MOCK
cat > "$tmp/bin/git" <<'MOCK'
#!/usr/bin/env bash
printf 'git %s\n' "$*" >> "$MOCK_LOG"
case $1 in
  clone) mkdir -p "$3"; printf '# README\n' > "$3/README.md" ;;
  config) exit 1 ;;
  diff) exit 1 ;;
  *) ;;
esac
MOCK
cat > "$tmp/bin/curl" <<'MOCK'
#!/usr/bin/env bash
url=${@: -1}
if [[ ${MOCK_CURL_503:-0} == 1 || $url == https://camo.* ]]; then printf '503 text/html 0'; else printf '200 image/png 3'; fi
MOCK
chmod +x "$tmp/bin/"*
export MOCK_LOG="$tmp/log" MOCK_BODY="$tmp/body"
script="$root/.claude/commands/scripts/attach-screenshots.sh"
out=$(PATH="$tmp/bin:$PATH" bash "$script" -i 1 -t 'Screenshots' "$tmp/png/foo-play.png" "$tmp/png/bar.png" 2>"$tmp/err")
[[ $out == https://github.com/testuser/project/issues/1\#issuecomment-987 ]]
grep -q '^git push$' "$MOCK_LOG"
[[ $(grep -c 'https://gist.githubusercontent.com' "$MOCK_BODY") -eq 2 ]]
grep -q '^### foo$' "$MOCK_BODY"
grep -q '^### bar$' "$MOCK_BODY"
grep -q 'gh gist create' "$MOCK_LOG"

: > "$MOCK_LOG"
PATH="$tmp/bin:$PATH" bash "$script" -i 1 -t Reused -g existing "$tmp/png/foo-play.png" "$tmp/png/bar.png" >/dev/null 2>/dev/null
! grep -q 'gh gist create' "$MOCK_LOG"

set +e
MOCK_CURL_503=1 PATH="$tmp/bin:$PATH" bash "$script" -i 1 -t Failure -g abc123 "$tmp/png/foo-play.png" "$tmp/png/bar.png" >/dev/null 2>"$tmp/503.err"
status=$?
set -e
[[ $status -eq 1 ]]
grep -q 'https://gist.githubusercontent.com/testuser/abc123/raw/foo-play.png' "$tmp/503.err"

# The proxied URL is what readers load; the direct URL answering 200 must not pass (catbox, 2026-10-14).
set +e
MOCK_CAMO=1 PATH="$tmp/bin:$PATH" bash "$script" -i 1 -t Camo -g abc123 "$tmp/png/foo-play.png" "$tmp/png/bar.png" >/dev/null 2>"$tmp/camo.err"
status=$?
set -e
[[ $status -eq 1 ]]
grep -q 'https://camo.githubusercontent.com/aa/foo' "$tmp/camo.err"

set +e
MOCK_COUNT_MISMATCH=1 PATH="$tmp/bin:$PATH" bash "$script" -i 1 -t Count -g abc123 "$tmp/png/foo-play.png" "$tmp/png/bar.png" >/dev/null 2>"$tmp/count.err"
status=$?
set -e
[[ $status -eq 1 ]]
grep -q 'image count: expected 2 got 1' "$tmp/count.err"

: > "$MOCK_LOG"
set +e
PATH="$tmp/bin:$PATH" bash "$script" -i 1 -t Missing -g abc123 "$tmp/png/nope.png" >/dev/null 2>"$tmp/missing.err"
status=$?
set -e
[[ $status -eq 2 ]]
! grep -Eq '^(gh|git) ' "$MOCK_LOG"
echo 'attach-screenshots: OK'
