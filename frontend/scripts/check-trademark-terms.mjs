#!/usr/bin/env bun
/**
 * Fails the build when a retired trademark reappears on a user-visible surface.
 *
 * The list of terms is NOT duplicated here — it is parsed out of the
 * `forbidden-terms` block in `TRADEMARKS.md`, which stays the single source of
 * truth. This repository's history is that a rule written in prose drifts away
 * from the code within a batch or two; making the document the guard's input is
 * what stops that.
 *
 * Why this guard exists: `21+3` was displayed in the BlackJack and Spanish 21
 * UI while being a live Japanese registration (Galaxy Gaming, 登録6752649). The
 * registration predated the feature by two years and four months, so nothing
 * about it was going to surface on its own — only a check at add time would
 * have caught it.
 *
 * Scope is deliberately user-visible strings only:
 *
 *   - i18n translation VALUES (keys are identifiers; `t3` must stay legal)
 *   - Go string literals that reach a presenter
 *   - game manuals under docs/manual/
 *
 * Identifiers, wire fields and code comments are out of scope: renaming
 * `twentyOnePlus3Bet` would break stored sessions and API clients for no
 * benefit, since no user ever sees it.
 */

import { readdirSync, readFileSync, statSync } from 'node:fs';
import { extname, join, relative } from 'node:path';
import { assertFloor } from './lib/floor.mjs';

/** Repository root — this file lives at frontend/scripts/. */
const REPO_ROOT = new URL('../../', import.meta.url).pathname;
const TRADEMARKS_MD = join(REPO_ROOT, 'TRADEMARKS.md');

/**
 * Sanity floor for scanned files. A walk that silently matches nothing exits 0
 * and reads exactly like a clean tree, which is how a license scanner fooled us
 * earlier in the same audit. Fail loudly instead.
 *
 * Raised from 200 to 4000: the walk covers 6212 files, so 200 was low enough that
 * losing `internal/` entirely — 5000-odd Go files, where most display strings
 * live — would still have cleared it. A floor has to sit under the real number,
 * not under zero.
 */
const MIN_FILES = 4000;

/** Directories scanned, each with the extensions that carry display strings. */
const TARGETS = [
  { dir: join(REPO_ROOT, 'frontend/src/i18n/locales'), exts: ['.json'] },
  { dir: join(REPO_ROOT, 'internal'), exts: ['.go'] },
  { dir: join(REPO_ROOT, 'docs/manual'), exts: ['.md'] },
];

/** Parses the forbidden-term list out of TRADEMARKS.md. */
function forbiddenTerms() {
  const md = readFileSync(TRADEMARKS_MD, 'utf8');
  const block = md.match(/<!-- forbidden-terms:start -->([\s\S]*?)<!-- forbidden-terms:end -->/);
  if (!block) {
    console.error('check-trademark-terms: no forbidden-terms block in TRADEMARKS.md.');
    process.exit(1);
  }
  const terms = [...block[1].matchAll(/^\s*-\s+`([^`]+)`/gm)].map((m) => m[1]);
  if (terms.length === 0) {
    console.error('check-trademark-terms: the forbidden-terms block is empty — expected at least one entry.');
    process.exit(1);
  }
  return terms;
}

/** Recursively lists files under `dir` with one of `exts`. */
function walk(dir, exts) {
  let out = [];
  let entries;
  try {
    entries = readdirSync(dir);
  } catch {
    console.error(`check-trademark-terms: cannot read ${dir} — the layout moved and this guard is scanning nothing.`);
    process.exit(1);
  }
  for (const entry of entries) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) out = out.concat(walk(full, exts));
    else if (exts.includes(extname(entry))) out.push(full);
  }
  return out;
}

/**
 * Yields the display strings of one file. For JSON this is the leaf values
 * only; for Go, double-quoted literals that are not on a comment line; for
 * Markdown, the whole document.
 */
function displayStrings(file) {
  const text = readFileSync(file, 'utf8');
  if (file.endsWith('.json')) {
    const values = [];
    const collect = (node) => {
      if (typeof node === 'string') values.push(node);
      else if (node && typeof node === 'object') for (const v of Object.values(node)) collect(v);
    };
    collect(JSON.parse(text));
    return values;
  }
  if (file.endsWith('.go')) {
    return text
      .split('\n')
      .filter((line) => !line.trimStart().startsWith('//'))
      .flatMap((line) => [...line.matchAll(/"((?:[^"\\]|\\.)*)"/g)].map((m) => m[1]));
  }
  return [text];
}

const terms = forbiddenTerms();
const violations = [];
let scanned = 0;

for (const { dir, exts } of TARGETS) {
  for (const file of walk(dir, exts)) {
    scanned += 1;
    let strings;
    try {
      strings = displayStrings(file);
    } catch (err) {
      console.error(`check-trademark-terms: cannot parse ${relative(REPO_ROOT, file)}: ${err.message}`);
      process.exit(1);
    }
    for (const term of terms) {
      if (strings.some((s) => s.includes(term))) {
        violations.push(`${relative(REPO_ROOT, file)}: contains "${term}"`);
      }
    }
  }
}

assertFloor('trademark-terms', scanned, MIN_FILES, 'files scanned');

if (violations.length > 0) {
  console.error(`check-trademark-terms: ${violations.length} user-visible use(s) of a retired trademark:`);
  for (const v of violations) console.error(`  - ${v}`);
  console.error('\nSee the forbidden-terms block in TRADEMARKS.md for the replacement wording.');
  process.exit(1);
}

console.log(
  `trademark-terms: OK (${scanned} files scanned, ${terms.length} retired term(s) absent from display strings).`,
);
