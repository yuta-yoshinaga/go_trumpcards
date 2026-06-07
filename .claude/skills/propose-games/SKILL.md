---
name: propose-games
version: 1.0.0
description: Propose NEW card games worth adding to go_trumpcards (deduped against the registry) and open each as a "<Game> の追加" GitHub issue. Use for "追加した方が良いゲームを提案", "新規ゲーム候補をissueに".
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
  - 追加した方が良いゲーム
  - 新規ゲームを提案
  - 追加候補のゲーム
  - propose new games
  - new game candidates
---

# propose-games — new-game candidates → GitHub issues

Sibling of [[game-improve]]. game-improve improves the 132 existing games;
**propose-games proposes games that are NOT yet implemented** and opens one
new-game issue per candidate. Codifies the 2026-06-07 batch (45 candidates).

## Output shape

One issue per candidate, titled `<English name> (<日本語名>) の追加`, body in the
repo's established new-game format (see issue #2013 Bourré as the canonical
example):

```
## 概要 (Overview)
## 背景と追加のメリット (Context & Why add this?)   (bullets: distinct mechanic / region / feasibility)
## 仕様案 (Requirements)                          (カテゴリ(バケット) / 人数 / デッキ / 主要ルール 4-6 steps)
## 実装のポイント (Implementation Details)          (reuse hint + new domain + checklist reminder)
```

## Decide scope first (AskUserQuestion unless already stated)

1. **候補数** — ~15 (curated) / ~25 / 40+ / 任せる.
2. **デッキ範囲** — standard decks only (52/54, Latin 40, pinochle 48) **or**
   special decks allowed (78-card tarot, 48-card hanafuda, etc.).

## Workflow

### 1. Enumerate existing games (dedup source of truth)
```sh
grep -oE '\{Name: "[^"]+"' internal/infrastructure/games/registry.go \
  | sed -E 's/.*"([^"]+)"/\1/' | sort > /tmp/newgames/existing.txt
```
This is the **only** authoritative dedup list — never propose a slug already in it.

### 2. Curate candidates (do this yourself — it's the value)
Pick real, notable card games across diverse buckets so the batch isn't lopsided:
trick-taking (regional: German Sheepshead/Doppelkopf, Spanish Mus/Tute, Italian
Scopone, Czech Mariáš…), rummy/melding (Hand and Foot, Conquian, Chinchón),
shedding/reflex (Mao, Spoons, Kemps), fishing (Pişti, Cuarenta), casino/vying
(Three Card Brag, Teen Patti, Five Card Stud, Faro, OFC), solitaire (Russian Bank,
La Belle Lucie, Simple Simon, Double Klondike, Black Hole), and — if special decks
are in scope — French Tarot, Koi-Koi, Go-Stop. Watch near-duplicates: `threecard`
= Three Card *Poker* (≠ Three Card Brag); `chinesepoker` = closed (≠ OFC); `scopa`
(≠ Scopone); `cassino` (≠ Pişti); `klondike` (≠ Double Klondike).

Write a seed JSON `/tmp/newgames/seeds.json` — array of
`{slug, jp, en, origin, deck, players, mechanic, reuse, bucket}`. `reuse` =
the existing game/component to model after (this is what makes proposals
implementable, not hand-wavy). `bucket` recommendation: **classic worker is at
the 1 MB gzip limit (#2126) → route new games to `casino` or `solo` only.**

Then verify zero collisions:
```sh
comm -12 <(jq -r '.[].slug' /tmp/newgames/seeds.json|sort -u) /tmp/newgames/existing.txt   # must be empty
```

### 3. Fan out body-writer agents (READ-ONLY)
Split seeds into groups of ~9; launch one `general-purpose` **sonnet** agent per
group in a single message. Each reads its `seedgroup-<n>.json`, expands every seed
into `{game,title,body}` following the template above, writes
`/tmp/newgames/group-<n>.json`. Enforce: **no build/test/lint/go/bun** (Read/Grep
/Glob only — OOM safety, see memory `feedback_sequential_tasks`); rules-accurate;
~180–280 word bodies; special decks (tarot/hanafuda) must call out a NEW non-52
Card/Deck domain type; social/real-time games (Mao/Spoons/Kemps) must describe a
concrete 1-human-vs-CPU adaptation.

### 4. Validate + merge
```sh
cd /tmp/newgames
jq -s 'add' group-*.json > all.json
jq 'length' all.json                                                       # == candidate count
jq -s 'add|map(select(.game==null or .title==null or .body==null))|length' group-*.json   # == 0
jq -s 'add|group_by(.game)|map(select(length>1))|length' group-*.json      # == 0
comm -12 <(jq -r '.[].game' all.json|sort) /tmp/newgames/existing.txt      # == empty (re-check dedup)
```
Spot-check one body for rule accuracy + the special-deck/social callouts.

### 5. Create issues
`FOOTER='...' SRC=/tmp/newgames/all.json bash scripts/create_issues.sh` (in this
skill dir). Idempotent (skips games already in `created.log`), `sleep 3` between
creates (GitHub secondary-rate-limit safety), `--body-file`, no `--label` (repo
has none). Run in background; ~3s/issue. For draft mode, render `all.json` as a
table and stop.

### 6. Verify + report + remember
```sh
comm -23 <(jq -r '.[].game' all.json|sort) <(cut -f1 created.log|sort)   # missing (want empty)
[ -s errors.log ] && cat errors.log
```
Report the issue-number range, per-bucket counts, and category spread. Record the
batch in project memory and add a MEMORY.md pointer.

## Notes
- Agent tool has no schema option (that's Workflow) — enforce JSON in the prompt,
  validate with jq.
- Do NOT use the Workflow tool unless the user explicitly opts into multi-agent
  orchestration ("ultracode" / "use a workflow").
- The hanafuda/tarot games imply a new deck abstraction — flag in the issue that
  they are larger efforts than a 52-card clone.
