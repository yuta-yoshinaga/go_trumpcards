---
name: game-improve
version: 2.0.0
description: Generate evidence-checked improvement proposals per card game and open them as GitHub issues. Use for "各ゲームの改善提案", "全ゲームのissueを作って", per-game UX-assist batches.
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

### 3. Build the evidence map
From the repository root, build `games.json`. The surface map records each
game's page, Web and CUI presenters, interactor, directly imported hooks and
components, transitive `deep` set (the full transitive closure of relative
imports under `frontend/src/components/` and `frontend/src/hooks/`), locales,
and `sync_cpu_loop` status.
It also adds up to eight matching closed issue titles per game, marking issues
closed as not planned.

```sh
python3 .claude/skills/game-improve/scripts/build_games.py --out games.json
```

### 4. Generate proposals in groups of 12
Generate one prompt per group with the script; it includes each game's mapped
surface, past issues, required proposal fields, and the known false positives.
Each proposal must include `current_state` entries (`path`, `line`, `note`) for
the behavior being criticized, plus `absence_evidence` entries (`claim`,
`pattern`, `paths`). The pattern describes code that would exist if the proposed
improvement were implemented. `current_state` was added after the smoke test
found patterns quoting existing code.
Use the same prompt text and JSON shape for either mode:

- When the session is in dele-mode:
  ```sh
  dele -k scan -s @.claude/skills/game-improve/scripts/schema.json "$(python3 .claude/skills/game-improve/scripts/mkprompt.py N --games games.json)"
  ```
- Otherwise, use read-only `Agent` subagents with that same prompt text. Validate
  their JSON shape with `jq` against the requested proposal structure.

Measured last time: codex took ~100–150 s per 12-game group; agy-pro took
250–650 s; Gemini ran out of quota after 9 groups.

### 5. Merge proposals and check absence claims
Merge the generated group JSON files (each shaped as `{"proposals":[...]}`) into `proposals.json`, then run the checker. It
validates required fields and evidence, searches every absence PCRE regex (as
`git grep -P`) over the
game's whole mapped surface including `deep` plus any paths named in the
evidence, and flags `needs_cpu_turn_state` for `sync_cpu_loop` games. Keep clean
proposals moving; send suspect proposals and their reported hits to step 6.

```sh
jq -s 'map(.proposals) | add' out-*.json > proposals.json
# A group can silently return fewer proposals than games (the smoke test dropped 1 of 3).
# Re-run the group for any game listed here before going on.
comm -23 <(jq -r '.[].game' games.json | sort) <(jq -r '.[].game' proposals.json | sort)
python3 .claude/skills/game-improve/scripts/check_absence.py proposals.json --games games.json --clean clean.json --suspect suspect.json
```

### 6. Verify suspect proposals and a clean sample
Render verification prompts with the suspect hits included. Review all suspect
proposals and a sample of clean ones; each verdict is `keep`, `fixed`, or
`replaced`. Re-run the absence checker on every fixed or replaced proposal:
replacements make new claims, and missed checks on replacements caused misses
last time.

```sh
python3 .claude/skills/game-improve/scripts/mkvprompt.py N --games games.json --proposals suspect.json --hits suspect.json
python3 .claude/skills/game-improve/scripts/mkvprompt.py N --games games.json --proposals clean.json
python3 .claude/skills/game-improve/scripts/check_absence.py fixed-and-replaced.json --games games.json --clean rechecked-clean.json --suspect rechecked-suspect.json
```

### 7. Cluster recurring premises
Run clustering after verification. A `premise_key` shared by at least three
distinct games becomes one policy issue listing those games, rather than
separate per-game issues. #8378 was closed per game before policy issue #8524
reversed that decision across 55 pages. Policy issues may need a user decision;
file them as such.

```sh
python3 .claude/skills/game-improve/scripts/cluster.py verified-proposals.json --min 3 --policy policy.json --individual individual.json
```

Handle policy candidates as decision items; create per-game issues only from
the individual proposals.

### 8. Create issues (idempotent, rate-limit-safe)
Use `scripts/create_issues.sh` (in this skill dir). It:
- creates one issue per `individual.json` element via `gh issue create --body-file`,
- appends a `対象ゲーム` footer,
- logs `game\turl\ttitle` to `created.log` and **skips already-logged games on
  rerun** (resumable if it dies mid-batch),
- `sleep 3` between creates (~20/min) to dodge GitHub's secondary
  content-creation rate limit. run in background.

Verify gh first: `gh auth status`. Choose labels from `gh label list` (for example,
`ui/ux`, `cli-ux`, or `refactor`).

Run the per-game issue command with `SRC=individual.json bash .claude/skills/game-improve/scripts/create_issues.sh`.

For draft mode: render the final proposal set to the user as a table and stop
before issue creation.

### 9. Verify + report
```sh
comm -23 <(jq -r '.[].game' individual.json|sort) <(cut -f1 created.log|sort)   # missing games (want empty)
wc -l created.log ; [ -s errors.log ] && cat errors.log
```
Report issue-number range, per-category counts, and the recurring-gap themes.
Record the batch in project memory (`project_issues_<lo>_<hi>_improvements`) and
add a one-line MEMORY.md pointer, mirroring `project_issues_2161_2292_improvements`.

## Why this pipeline (measured, #8071–#8453)

Of 383 filed proposals, 82 (21%) were closed as not planned because the premise
was wrong. This was after the second pass had already replaced 269 of the 383
first-pass proposals.

The dominant cause was that the feature already existed outside the page file:
shared components such as `VideoPokerGameContent` and `CardImage`, the Web
presenter's message, the CUI presenter, and always-mounted live regions. Both
passes read the same two files, so they shared the blind spot.

About 5 proposals asked to show “the CPU's turn” in games whose interactor runs
CPU turns synchronously (Vint / TwoTenJack / Tute / Rikken / Pineapple).
A replay of real false positives through `check_absence.py` caught Deuces Wild
(via `VideoPokerGameContent` only), Gin Rummy (via `CardImage`), Horse (CUI
presenter), and Tute (pattern plus sync loop).

A generic pattern such as `aria-live` matches shared components everywhere, so
vague evidence is pushed back rather than passed. That is intended.

## Known false positives

- Keyboard shortcuts come from shared hooks / `useActionShortcuts` — shared hook implementation.
- `CardImage` already provides a default alt through `cardAlt` — `CardImage` component.
- The video-poker family renders through `VideoPokerGameContent` — `VideoPokerGameContent` component.
- Live regions are often always-mounted sr-only elements with a testid — page-level live-region elements.
- Column numbers are 0-based by repository convention; never propose 1-based numbering — game column/index logic.
- Purely decorative motion (screen shake and similar) is low value — presentation-only motion code.
- In `sync_cpu_loop` games, never propose showing the CPU's turn — interactor synchronous CPU loop.
- Do not treat a config value as a play statistic (Whitehead `drawCount`) — Whitehead config value.

## Notes / gotchas
- The generation step must follow dele-mode when it is on. A skill telling the
  model to fan out agents once overrode dele-mode and burned ~6.6M tokens.
- `Agent` tool has **no schema option** (that's Workflow). Enforce JSON shape in
  the prompt and validate with `jq` after.
- Do NOT use the Workflow tool here unless the user explicitly opts into
  multi-agent orchestration ("ultracode" / "use a workflow").
- Frontend page = `frontend/src/pages/`, CUI presenter = `internal/adapter/presenter/`.
- Body files are reused per-iteration at `/tmp/claude-proposals/body.md`; the
  script overwrites it each loop — safe because it writes-then-creates serially.
