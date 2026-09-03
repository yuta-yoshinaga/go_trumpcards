#!/usr/bin/env bun
// Guard that every bid/discard prompt renders inside an **always-mounted**
// live region.
//
// The prompt is text that appears when the phase changes. Without
// `role="status"` / `aria-live` nothing reaches a screen reader, so the turn
// passes to the player silently. Twelve pages were in that state when this
// guard was written (#6880, following the same shape #6484 established for
// Calabresella).
//
// The subtle half is the same one `check-hint-live-region.mjs` documents: a
// region that enters the DOM already holding its text is generally not
// announced, because assistive tech watches an *existing* region for changes.
// Putting the attributes on the `{canBid && (<div>…</div>)}` that only exists
// during the bid phase therefore passes an "aria-live is present" check while
// still announcing nothing. This guard rejects that and requires the region to
// be mounted unconditionally, with only its contents swapping.
//
// It also rejects a region that merely *sits beside* the prompt: the prompt has
// to be a descendant, which is what `toContainElement` asserts in the page
// tests.

import { readdir, readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const PAGES = process.argv[2] ?? fileURLToPath(new URL('../src/pages', import.meta.url));

/** Prompts this guard governs. Other prompt kinds ride along when adjacent. */
const PROMPT_RE = /data-testid="([a-z0-9]+-(?:bid|discard)-prompt)"/g;

/** The permanent wrapper each page must provide. */
const REGION_RE = /<div\s+data-testid="([a-z0-9]+-prompt-live)"([^>]*)>/;

/**
 * Whether the element opened at `at` is gated by a conditional.
 *
 * @param {string} src - Page source.
 * @param {number} at - Offset of the opening tag.
 * @returns {boolean} True when `&&` or `?` immediately precedes it.
 */
function isConditionallyMounted(src, at) {
  return /(?:&&|\?)\s*\(?\s*$/.test(src.slice(Math.max(0, at - 80), at));
}

/**
 * Offset just past the element opened at `open`, matching nested `<div>`s.
 *
 * @param {string} src - Page source.
 * @param {number} open - Offset of the region's opening tag.
 * @returns {number} Offset of the matching close, or -1.
 */
function closeOf(src, open) {
  const tags = [...src.slice(open).matchAll(/<\/?div\b[^>]*>/g)];
  let depth = 0;
  for (const t of tags) {
    if (t[0].startsWith('</')) {
      depth -= 1;
      if (depth === 0) return open + t.index + t[0].length;
    } else if (!t[0].endsWith('/>')) depth += 1;
  }
  return -1;
}

const problems = [];
let pages = 0;
let prompts = 0;

for (const name of (await readdir(PAGES)).sort()) {
  if (!name.endsWith('Page.tsx')) continue;
  const src = await readFile(join(PAGES, name), 'utf8');
  const found = [...src.matchAll(PROMPT_RE)];
  if (found.length === 0) continue;
  pages += 1;
  prompts += found.length;

  const region = REGION_RE.exec(src);
  if (region === null) {
    problems.push(`${name}: ${found.map((m) => m[1]).join(', ')} has no *-prompt-live region`);
    continue;
  }
  const attrs = region[2];
  if (!/role="status"/.test(attrs) || !/aria-live=/.test(attrs)) {
    problems.push(`${name}: ${region[1]} needs both role="status" and aria-live`);
    continue;
  }
  if (isConditionallyMounted(src, region.index)) {
    problems.push(
      `${name}: ${region[1]} is mounted conditionally. A region that appears together with its text is not announced; mount it always and swap the contents.`,
    );
    continue;
  }
  const end = closeOf(src, region.index);
  for (const m of found) {
    if (m.index < region.index || m.index > end) {
      problems.push(`${name}: ${m[1]} is outside ${region[1]}, not inside it`);
    }
  }
}

assertFloor('prompt-live-region', pages, 8, 'pages with a bid/discard prompt');
assertFloor('prompt-live-region', prompts, 12, 'bid/discard prompts scanned');

if (problems.length > 0) {
  console.error('prompt-live-region: NG');
  for (const p of problems) console.error(`  ${p}`);
  process.exit(1);
}
console.log(`prompt-live-region: OK (${prompts} prompts across ${pages} pages).`);
