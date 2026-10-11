---
name: term-unify
description: Safely unify terminology across a complete game in CUI, Web, manual, and CLI formatter surfaces. Use for 「用語を統一」, 「訳語を揃えて」, or “rename a term across games”.
argument-hint: "<game> <old term> <new term>"
---

# Term Unify

Use this workflow for a terminology change such as replacing 「場札」 with 「組札」. Work on one **whole game** at a time. Never split the work by surface: change that game's CUI (Go presenter and Go locale), Web (frontend locale and pages), manual, and CLI formatter (`frontend/src/utils/cli/formatters`) together.

## Before editing

1. Confirm the requested game and exact old/new strings.
2. Read the target game's existing terminology in every surface, plus nearby rows and examples. Identify whether the term appears inside parenthetical explanations.
3. Count the baseline on `HEAD`, never the worktree. This captures both the old and new spelling and catches pre-existing `new（` compounds:

   ```bash
   bash .claude/skills/term-unify/scripts/before.sh '旧表記' '新表記' path/to/game/files...
   ```

   `git grep -c` counts matching lines per file. Record all three totals in the change notes. Check the surrounding lines for usage counts that line totals cannot express.
4. Inventory all surfaces for this game and plan to update each one in this task. Do not stop after Web, manual, or CUI alone.

## Make the change

Replace only the standalone term in the intended display text. Preserve parenthetical clarifications and punctuation; review each changed line in context. Keep neighboring entries and examples consistent. Do not use a broad replacement that can produce `新語（新語）`.

## Verify before finishing

Run the after-check against every modified locale file (and other changed files if they contain the term):

```bash
bash .claude/skills/term-unify/scripts/after.sh '旧表記' '新表記' path/to/changed/files...
```

It fails if a Japanese locale contains `同語（同語）`, if any changed target file has additions different from deletions in `git diff --numstat`, if the old spelling remains, if the new spelling count falls below its `HEAD` baseline, or if the count of `新語（` exceeds its `HEAD` baseline. A missing target file is not evidence of success: inspect the diff and verify every intended line and file remains present. Review `git diff --check` and the full scoped diff.

If any locale file changed, run the required Go tests even when the Go diff is empty:

```bash
go test -tags test ./...
```

This is a project requirement for locale edits. Also run the unit test for the term-unify scripts:

```bash
bash .claude/hooks/term-unify-scripts.test.sh
```

Report the game, all four surfaces checked, the three before/after counts, line preservation, repetition result, and verification commands with their actual outcomes.
