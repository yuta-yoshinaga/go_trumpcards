---
name: game-improve
version: 1.0.0
description: Generate ONE high-impact improvement proposal per card game (from real Page.tsx + CuiPresenter code) and open each as a GitHub issue. Use for "各ゲームの改善提案", "全ゲームのissueを作って", per-game UX-assist batches.
allowed-tools:
  - Bash
  - Read
  - Write
  - Edit
  - Grep
  - Glob
  - Agent
  - AskUserQuestion
triggers:
  - 各ゲームの改善提案
  - ゲームの改善提案をissueに
  - 全ゲームのissue
  - per-game improvement issues
  - game improvement batch
---

# game-improve — per-game improvement proposals → GitHub issues

Generate one code-grounded improvement proposal per in-scope game and open it as a GitHub issue.

## What it produces

For each in-scope game: ONE highest-impact, **code-grounded** improvement
(named files / phases / components — never generic "add tests"), opened as a
GitHub issue with the body template:

```
## 背景・現状
## 提案
## 受け入れ条件   (- [ ] checklist)
## 対象ファイル   (`path` list)
```

Titles are Japanese with a bracket tag: `[UI/UX]` `[機能]` `[アクセシビリティ]`
`[CUI]` `[i18n]` + the game's Japanese name in parentheses.

## Decide scope first (AskUserQuestion unless the user already said)

1. **対象範囲** — all games / one category / explicit list /
   only games without an existing improvement issue.
2. **提案数/game** — default 1 (most effective). 2–3 → one issue each.
3. **作成タイミング** — immediate (`gh issue create` now) vs draft-then-confirm.
4. (Skip skillify question — this *is* the skill.)

## Workflow

### 1. Canonical game list (SSoT)
Read `internal/infrastructure/games/registry.go` (the SSoT). Each entry is
`{Name, Category, Description}`. Name = short slug + URL segment; Description
carries the Japanese name. Get the current game list with
`go run ./cmd/trumpcards games --short`.

### 2. Dedupe against open issues
`gh issue list --state open --limit 400 --json number,title`. Skip games that
already have a comparable open improvement issue. (Most past per-game issues are
closed/merged, so collisions are rare.)

### 3. Fan out READ-ONLY analysis agents
Chunk the in-scope games into groups of ~16. Launch one `general-purpose`
**sonnet** agent per group **in a single message** (parallel). Per-agent prompt
must enforce:

- Agents are read-only (Read/Grep/Glob) — this is a proposal pass, not an implementation.
- For each game: read `frontend/src/pages/<PascalCase>Page.tsx` and
  `internal/adapter/presenter/<PascalCase>CuiPresenter.go` (glob for casing);
  optionally the hook/components or `docs/manual/web/<game>.md`.
- Output ONE proposal/game (or N if scope said so) to
  `/tmp/claude-proposals/group-<n>.json` as a JSON array of
  `{game,title,body}`. Body ~120–220 words, valid JSON (escape newlines).
- Return only a one-line count; the file is the real payload (keeps the
  orchestrator's context lean).

Map PascalCase from the slug if needed (e.g. `bigtwo`→`BigTwo`,
`seahaventowers`→`SeahavenTowers`, `ohhell`→`OhHell`).

### 4. Validate + merge
```sh
cd /tmp/claude-proposals
for f in group-*.json; do jq -e . "$f" >/dev/null && echo "$f OK $(jq length $f)" || echo "$f BAD"; done
jq -s 'add' group-*.json > all.json
jq 'length' all.json                                   # == in-scope count
jq -s 'add|map(select(.game==null or .title==null or .body==null))|length' group-*.json   # == 0
jq -s 'add|group_by(.game)|map(select(length>1))|map(.[0].game)' group-*.json             # == []
```
Spot-check one full body for quality (real file names, concrete phase).

### 5. Create issues (idempotent, rate-limit-safe)
Use `scripts/create_issues.sh` (in this skill dir). It:
- creates one issue per `all.json` element via `gh issue create --body-file`,
- appends a `対象ゲーム` footer,
- logs `game\turl\ttitle` to `created.log` and **skips already-logged games on
  rerun** (resumable if it dies mid-batch),
- `sleep 3` between creates (~20/min) to dodge GitHub's secondary
  content-creation rate limit. run in background.

Verify gh first: `gh auth status`. Choose labels from `gh label list` (for example,
`ui/ux`, `cli-ux`, or `refactor`).

For draft mode: render `all.json` to the user as a table and stop before step 5.

### 6. Verify + report
```sh
comm -23 <(jq -r '.[].game' all.json|sort) <(cut -f1 created.log|sort)   # missing games (want empty)
wc -l created.log ; [ -s errors.log ] && cat errors.log
```
Report issue-number range, per-category counts, and the recurring-gap themes.
Record the batch in project memory (`project_issues_<lo>_<hi>_improvements`) and
add a one-line MEMORY.md pointer, mirroring `project_issues_2161_2292_improvements`.

## Notes / gotchas
- Agent tool has **no schema option** (that's Workflow). Enforce JSON shape in the
  prompt and validate with `jq` after.
- Do NOT use the Workflow tool here unless the user explicitly opts into
  multi-agent orchestration ("ultracode" / "use a workflow"). Plain `Agent`
  fan-out is the default.
- Frontend page = `frontend/src/pages/`, CUI presenter = `internal/adapter/presenter/`.
- Body files are reused per-iteration at `/tmp/claude-proposals/body.md`; the
  script overwrites it each loop — safe because it writes-then-creates serially.
