#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
Usage: attach-screenshots.sh -i <issue-number> -t <heading> [-g <gist-id>] [-R <owner/repo>] <png>...
EOF
}

die_usage() { usage; exit 2; }
issue='' heading='' gist='' repo=''
while getopts ':i:t:g:R:' opt; do
  case $opt in
    i) issue=$OPTARG ;;
    t) heading=$OPTARG ;;
    g) gist=$OPTARG ;;
    R) repo=$OPTARG ;;
    *) die_usage ;;
  esac
done
shift $((OPTIND - 1))
[[ -n $issue && -n $heading && $# -gt 0 ]] || die_usage
for png in "$@"; do [[ -f $png ]] || { echo "PNG does not exist: $png" >&2; die_usage; }; done
command -v gh >/dev/null && gh auth status >/dev/null 2>&1 || { echo 'gh must be logged in' >&2; die_usage; }
command -v git >/dev/null || { echo 'git is required' >&2; die_usage; }
command -v curl >/dev/null || { echo 'curl is required' >&2; die_usage; }
owner=$(gh api user -q .login)
[[ -n $repo ]] || repo=$(gh repo view --json nameWithOwner -q .nameWithOwner)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

if [[ -z $gist ]]; then
  printf '# UX review screenshots\n\nScreenshots attached to UX review issues.\n' > "$tmp/README.md"
  gist=$(gh gist create "$tmp/README.md" 2>/dev/null | sed -E 's#.*/##')
  [[ -n $gist ]] || { echo 'could not create gist' >&2; exit 1; }
  echo "gist: $gist" >&2
fi
git clone "https://gist.github.com/${gist}.git" "$tmp/gist" >/dev/null 2>&1
cd "$tmp/gist"
for png in "$@"; do cp "$png" "$(basename "$png")"; done
git add -- .
if ! git diff --cached --quiet; then
  name=$(git config user.name || true); email=$(git config user.email || true)
  if [[ -z $name || -z $email ]]; then
    login=$(gh api user -q .login)
    name=${name:-$login}; email=${email:-"${login}@users.noreply.github.com"}
  fi
  git -c user.name="$name" -c user.email="$email" commit -m 'Add UX review screenshots' >/dev/null
  git push >/dev/null
fi

body="$tmp/body.md"
printf '## %s\n' "$heading" > "$body"
for png in "$@"; do
  base=$(basename "$png")
  label=${base%-play.png}; label=${label%.png}
  printf '\n### %s\n![%s](https://gist.githubusercontent.com/%s/%s/raw/%s)\n' "$label" "$label" "$owner" "$gist" "$base" >> "$body"
done
comment_url=$(gh issue comment "$issue" --repo "$repo" --body-file "$body")
comment_id=${comment_url##*issuecomment-}
[[ $comment_id =~ ^[0-9]+$ ]] || { echo "could not parse comment id from: $comment_url" >&2; exit 1; }
html=$(gh api "repos/${repo}/issues/comments/${comment_id}" -H 'Accept: application/vnd.github.html+json' -q .body_html)
# Read the src attribute itself: for third-party hosts GitHub points src at its camo proxy and moves the
# original URL to data-canonical-src, and only the proxied URL is what readers actually load.
mapfile -t urls < <(printf '%s' "$html" | grep -oE '<img[^>]*>' | sed -nE 's/^<img([^>]*[[:space:]])?src="([^"]+)".*/\2/p')
failures=()
if (( ${#urls[@]} != $# )); then failures+=("image count: expected $# got ${#urls[@]}"); fi
for url in "${urls[@]}"; do
  ok=0
  for attempt in 1 2 3; do
    result=$(curl -sS -L -o /dev/null -w '%{http_code} %{content_type} %{size_download}' "$url" 2>/dev/null || true)
    read -r status type size <<< "$result"
    if [[ $status == 200 && $type == image/* && $size =~ ^[0-9]+$ ]] && (( size > 0 )); then ok=1; break; fi
    (( attempt == 3 )) || sleep 2
  done
  (( ok )) || failures+=("$url (last response: ${result:-none})")
done
if (( ${#failures[@]} )); then printf 'Screenshot verification failed: %s\n' "${failures[@]}" >&2; exit 1; fi
printf '%s\n' "$comment_url"
