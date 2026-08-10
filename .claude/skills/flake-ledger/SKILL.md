---
name: flake-ledger
description: Record a test failure and decide whether it is actually a flake, by counting sightings at the same location instead of trusting a green re-run. Use when a test fails and you are tempted to re-run it, when CI is red on a change that could not have caused it, or to review which failures have repeated ("this test is flaky, right?", "re-run CI").
# Model-invocable: read-only bookkeeping. Its whole point is to fire at the moment a
# test fails and a re-run is tempting — a human trigger arrives too late to help.
allowed-tools: Read, Grep, Glob, Bash
argument-hint: "[record|check|list] [test-id]"
---

# Flake Ledger

**"I re-ran it and it passed" is not evidence of a flake.** It is one observation with nothing
to compare it to — a real, intermittent bug produces exactly the same reading. The only thing
that distinguishes a flake from a bug you got lucky on is *the same failure happening at the
same location more than once*, and that requires writing the first one down.

This skill writes it down.

## When a test fails

### 1. Read the actual result before naming it

A failure has to be read from the run's own output, not from a wrapper's exit status:

- **Background runs**: the completion notification's exit code is the code of the *echo* at the
  end of the command, not the test run. Read the `EXIT` line in the log file.
- **Vitest shards**: a worker that dies on a heap limit skips the tests queued behind it and
  the run can still report green. Check the reported test count, not just the colour.
- **Guard scripts**: an empty scan exits 0. A guard that printed no findings has not
  necessarily checked anything (see `frontend/scripts/lib/floor.mjs`).

### 2. Record the sighting

```bash
python3 .claude/skills/flake-ledger/scripts/ledger.py record \
  --test 'TestKaiserAllPassRedeal' \
  --where 'internal/domain' \
  --run 'https://github.com/<owner>/<repo>/actions/runs/<id>' \
  --note 'redeal loop asserted 4 hands, got 3'
```

`--where` is the package or file, and it matters: two different tests failing in the same
package is a different signal from one test failing twice. Omit `--run` for a local run.

The command prints the sighting count at that location and, with it, the verdict:

| Sightings | Verdict | What to do |
|---|---|---|
| 1 | UNCONFIRMED | **Investigate as a real failure.** Do not re-run and move on. |
| 2+ | CONFIRMED FLAKE | Open or update a GitHub issue with all the recorded runs. |

### 3. Act on the verdict, not on the re-run

On UNCONFIRMED, the next step is reading the test and the code under it — the same work you
would do for a reproducible failure. Suspect your own branch before the suite: check whether
your diff touches the failing package at all (`git diff --stat develop...HEAD -- <pkg>`).

On CONFIRMED, the ledger rows are the issue body. Paste them; a flake report with two dated
runs gets fixed, one that says "this is flaky sometimes" does not.

## Reviewing the ledger

```bash
python3 .claude/skills/flake-ledger/scripts/ledger.py check          # everything, grouped
python3 .claude/skills/flake-ledger/scripts/ledger.py check --test TestX
python3 .claude/skills/flake-ledger/scripts/ledger.py list --since 2026-07-01
```

`check` splits the ledger into CONFIRMED (repeated at one location) and UNCONFIRMED
(single sightings, which are still open questions rather than closed ones).

## Where the ledger lives

`.claude/.flake-ledger.jsonl`, gitignored — it is local observation, not shared truth. The
shared artefact for a confirmed flake is a **GitHub issue**. Do not promote the ledger into a
committed file; it would go stale the moment a flake is fixed and nobody deleted the row.

## Known-flaky background

Before recording, check whether the failure is already a known one — the repo has a running
list of backend and E2E flakes, and re-litigating a known flake wastes the investigation. If
it is already known, still record the sighting: the count is what tells you whether a "known
flake" is getting worse.
