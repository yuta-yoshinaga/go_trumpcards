#!/usr/bin/env bun
// Guard that every Markdown table row has the same number of cells as its
// header separator.
//
// Two ways a row breaks, both of which render as garbage and neither of which
// any existing check looks at:
//
//   1. **A merge collapses two rows into one.** `frontend/CLAUDE.md` had
//      `… | First-visit dialog … hook || \`TutorialProgressPanel\` | … |` — one
//      line holding two rows. It survived because the `useGameHint` row directly
//      above it carries the hint-factory count, which conflicts on nearly every
//      parallel PR (#4652), so that spot gets machine-resolved constantly. tsc,
//      Biome and vitest all ignore `.md`, so nothing caught it; it was found by
//      eye, and by then a *third* row had been lost entirely.
//
//   2. **An unescaped pipe inside a cell.** `docs/manual/cui/dragontiger.md`
//      documented `b <amount> <d|t|e>`, and those pipes split the row into five
//      cells. The other 30-odd manuals escape them (`<d\|t\|e>`); these two did
//      not.
//
// The check is deliberately shallow — cell counts only, no rendering — because
// that is enough to catch both shapes and cannot produce opinions about style.

import { readdir, readFile } from 'node:fs/promises';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const REPO = join(FRONTEND, '..');

/** Directories that are not ours to police. */
const SKIP = new Set(['node_modules', '.git', 'dist', 'coverage', 'playwright-report', 'test-results']);

/** Count unescaped pipes — `\|` is a literal inside a cell, not a separator. */
function cellCount(line) {
  return (line.match(/(?<!\\)\|/g) ?? []).length;
}

/** A separator row: only dashes, colons, pipes and spaces, with real dashes. */
function isSeparator(line) {
  const stripped = line.replace(/[|\s]/g, '');
  return stripped.length > 0 && /^[-:]+$/.test(stripped) && line.includes('---');
}

async function* markdownFiles(dir) {
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    if (SKIP.has(entry.name)) continue;
    const full = join(dir, entry.name);
    if (entry.isDirectory()) yield* markdownFiles(full);
    else if (entry.name.endsWith('.md')) yield full;
  }
}

const problems = [];
let tables = 0;
let rows = 0;

for await (const file of markdownFiles(REPO)) {
  const lines = (await readFile(file, 'utf8')).split('\n');
  let expected = null;
  for (const [i, line] of lines.entries()) {
    if (!line.trimStart().startsWith('|')) {
      expected = null;
      continue;
    }
    const cells = cellCount(line);
    if (isSeparator(line)) {
      expected = cells;
      tables += 1;
      continue;
    }
    if (expected === null) continue;
    rows += 1;
    if (cells !== expected) {
      problems.push(
        `${relative(REPO, file)}:${i + 1} — ${cells} cell separators, header has ${expected}\n      ${line.trim().slice(0, 90)}`,
      );
    }
  }
}

if (tables === 0) {
  console.error('markdown-tables: found no tables at all — the separator match has drifted.');
  process.exit(1);
}

if (problems.length > 0) {
  console.error('Markdown table rows that do not match their header:\n');
  for (const p of problems) console.error(`  ${p}`);
  console.error('\nA row split by a merge, or a `|` inside a cell that needs escaping as `\\|`.');
  process.exit(1);
}

console.log(`markdown-tables: OK (${rows} rows across ${tables} tables).`);
